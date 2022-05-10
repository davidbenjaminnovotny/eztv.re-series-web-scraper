
# EZTV.RE SERIES WEB SCRAPER

Written in Golang and with the help of Colly, this web scraper returns all Series and their torrent children to either a JSON or direct into your SQL DB.

## Are you ready to "Go"?
Make sure you have Go set up.   
[Get it here!](https://go.dev/doc/install)

## Go lang or go home

#### 1. Set up your Env to match your SQL DB.
Dont use these settings they well cause the lizzard people of old to come back from the moon.
So change the values to your own.
```
DB_USER=USERPOO
DB_PASSWORD=PASSYDOO
DB_HOST=HOSTIPAH
DB_DATABASE=DATABASITA
DB_PORT=PORTYAHHHHH

```

#### 2. Make sure you got your dependencies in order.
Try:
```
go mod tidy

```
If that doesnt work download each individual 3rd party dependency listed in the import in the main.go file.

Example: 
```
go get github.com/gocolly/colly

```
#### 3. Just run it.

```
go run . 

```

or 
```
go run main.go

```

