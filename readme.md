## Description
gateway is a restful API example server in golang. it shows how to pull data from database and cache it in the redis for the better performance 

    Usage: gateway [options]
           
    Options:
      -dbDriver string
            the name of database driver
      -dbSource string
            the name of database source string
      -hostAddress
            the server address

By default it is running like: gateway -dbDriver=sqlite3 -dbSource=cab.db -hostAddress=0.0.0.0:3000

## Prerequisite 

It requres redis server running locally

if it is unavailiable, run "sudo apt-get install redis-server" (Ubuntu) to set up redis server

## Build Instructions (Linux only)

Run the commands in sequence:  

go get -d

go build

## Tests
go test
go test ./services

## Command line for testing

install the tool 'curl' locally and run the commands as below

1) command to get a list of unique medallion:  

curl -i -GET "localhost:3000/api/medallion"

2) command to get the trip details for some medallions given a particular pickup date:

curl -i -POST "localhost:3000/api/trips" "{'medallions':[], 'date':''}"

3) command to ignore cache to get fresh result:

curl -i -POST "localhost:3000/api/trips" "{'medallions':[], 'date':'','fresh':true}"

4) command to clear the cache

curl -i -DELETE "localhost:3000/api/cache"
