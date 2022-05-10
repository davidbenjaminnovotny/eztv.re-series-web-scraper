package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Torrent Object/Struct
type Torrent struct {
	Title   string `json:"title"`
	Hash    string `json:"link"`
	Size    string `json:"size"`
	Seeders string `json:"seeders"`
}

// Series Object/Struct
type Series struct {
	Title    string    `json:"title"`
	Link     string    `json:"link"`
	Torrents []Torrent `json:"torrents"`
}

// Save Series & and related torrents to separate tabels in database
// This is a temporary solution ( Performance can still be heavilty improved )

// Call to .env file
func getEnvVar(key string) string {

	// Load environment variables & Check for errors
	if err := godotenv.Load(); err != nil {
		logrus.Error(err)
	}

	// Return the environment variable
	return os.Getenv(key)
}

func saveToDb(series []Series) {

	// Open Con using preset ENV variables
	db, err := sql.Open("mysql", getEnvVar("DB_USER")+
		":"+getEnvVar("DB_PASSWORD")+
		"@tcp("+getEnvVar("DB_HOST")+
		")/"+getEnvVar("DB_DATABASE"))

	// Check Con for errors
	if err != nil {
		logrus.Error(err)
	}

	// Close and Consolidate Connection
	defer db.Close()

	// Insert Series into existing table
	stmt, err := db.Prepare("INSERT INTO eztv_series (title, link) VALUES (?, ?)")

	// Check for prepare errors
	if err != nil {
		logrus.Error(err)
	}

	// Execute the statement in loop
	for _, serie := range series {
		res, err := stmt.Exec(serie.Title, serie.Link)
		if err != nil {
			logrus.Error(err)
		}
		// Take the last insert id
		id, err := res.LastInsertId()

		// check for errors in last insert id
		if err != nil {
			println("Error:", err.Error())
		} else {

			// Prepare the statement to insert torrents into existing table
			stmt, err := db.Prepare("INSERT INTO eztv_series_sources (torrent_id, title, link, size, seeders) VALUES (?, ?, ?, ?, ?)")

			// Check for prepare errors
			if err != nil {
				logrus.Error(err)
			}

			// Execute the statement in loop
			for _, serie := range serie.Torrents {
				_, err := stmt.Exec(id, serie.Title, serie.Hash, serie.Size, serie.Seeders)

				// Check for errors in table insert execution
				if err != nil {
					logrus.Error(err)
				}
			}

		}

	}
}

// Check if file exists in directory
func checkFile(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

// writes series to json file ( for testing purposes )
func writeJSON(data []Series) {

	// designate file to write to
	filename := "series.json"

	// Check the designated file
	err := checkFile(filename)

	// Check the file for errors
	if err != nil {
		logrus.Error(err)
	}
	// Read the file
	file, err := ioutil.ReadFile(filename)
	// Check for errors
	if err != nil {
		logrus.Error(err)
	}

	// Create a temporary Struct
	newStruct := []Series{}

	// Marshal the file into the temporary struct
	json.Unmarshal(file, &newStruct)

	// Append the new series to the temporary struct
	newStruct = append(newStruct, data...)

	// Marshal the temporary struct into a byte array
	dataBytes, err := json.MarshalIndent(newStruct, "", " ")

	// Check for errors
	if err != nil {
		logrus.Error(err)
	}

	// Write the byte array to the file
	err = ioutil.WriteFile(filename, dataBytes, 0644)

	// Check for errors
	if err != nil {
		logrus.Error(err)
	}
}
func main() {
	// Create a new Collector
	c := colly.NewCollector(
		colly.AllowedDomains("eztv.re"),
	)

	// Gather the series links
	c.OnHTML("tbody tr td:nth-child(1) .thread_link", func(e *colly.HTMLElement) {

		// Collect the series link
		link := e.Attr("href")

		// Visit link of the series
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// fires when a series is visited
	c.OnRequest(func(r *colly.Request) {
	})

	// Access table on series directory page to gather data
	c.OnHTML("table:nth-child(4) ", func(e *colly.HTMLElement) {

		// Create a new temporary struct
		tempSeries := []Series{}

		// Target Title of Series
		titleChange := e.ChildText("tbody tr:nth-child(1) td:nth-child(1) h2")

		// Remove redundant characters from title
		titleChange = strings.ReplaceAll(titleChange, "Torrent Download", "")

		// Add Series data to struct
		serie := Series{
			Title:    titleChange,
			Link:     e.Request.URL.String(),
			Torrents: []Torrent{},
		}

		// Loop through rows to gather series torrents data
		e.ForEach("tr", func(i int, e *colly.HTMLElement) {
			serie.Torrents = append(serie.Torrents, Torrent{
				Title:   e.ChildText(".epinfo"),
				Hash:    e.ChildAttr(".magnet", "href"),
				Size:    e.ChildText("td:nth-child(4)"),
				Seeders: e.ChildText("td:nth-child(6)"),
			})
		})

		// remove first two redundant elements in torrent results
		serie.Torrents = append(serie.Torrents[:0], serie.Torrents[2:]...)

		// append the series to temporary struct
		tempSeries = append(tempSeries, serie)

		// Begin process to insert series & torrents into database
		saveToDb(tempSeries)

		// Write the series to json file
		// writeJSON(tempSeries)
	})

	// Vists the series directory page on eztv.re
	c.Visit("https://eztv.re/showlist/")
}
