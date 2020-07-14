## Description
gateway is a restful API example server in golang. it shows how to pull data from database and cache it for the better performance 

    Usage: gateway [options]
           
    Options:
      -dbDriver string
            the name of database driver
      -dbSource string
            the name of database source string
      -hostAddress
            the server address

By default it is running like: gateway -dbDriver=sqlite3 -dbSource=./services/cab.db -hostAddress=0.0.0.0:3000

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

curl -i -X GET "localhost:3000/api/medallion"

2) command to get the trip details for some medallions given a particular pickup date:

curl -i -X POST "localhost:3000/api/trips" -d "{\"medallions\":[\"11B683D81211E96F472173C69D4632D6\",\"9A80FE5419FEA4F44DB8E67F29D84A0F\"], \"date\":\"2013-12-31\"}"

3) command to ignore cache to get fresh result:

curl -i -X POST "localhost:3000/api/trips" -d "{\"medallions\":[\"11B683D81211E96F472173C69D4632D6\",\"9A80FE5419FEA4F44DB8E67F29D84A0F\"], \"date\":\"2013-12-31\",\"fresh\":true}"


4) command to clear the cache

curl -i -X DELETE "localhost:3000/api/cache"
