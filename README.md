# Nommer

API for ingesting events/logs/any data into a Mongo Database categorized by project.

# Setup

You can clone this repo, or pull the docker image:

```bash
docker pull cjhammons/nommer:latest
```

You will need a .env file with your MongoDB URI and database, like this:

```
MONGO_URI=mongodb://<ip address>
MONGO_DATABASE=<database name>
```

# Usage

To start sending data to the API, you'll first need to create a project via the POST /project endpoint. Then you can send data to this project via POST /{project_name}/event using the API key generated when the project is created.

## Endpoints

### POST /1/projects

Creates a project and returns an API key.

Example:
```bash
curl -X POST "http://localhost:8080/1/projects" \
     -H "Content-Type: application/json" \
     -d '{"name": "MyNewProject"}'
```
Be sure to write down the apikey!

### POST /1/{project_name}/event

Sends an event to the project.

Example:
```bash
curl -X POST "http://localhost:8080/1/test_project/event" \
     -H "Content-Type: application/json" \
     -H "X-API-Key: your_actual_api_key_here" \
     #Note the data can take ANY form, as long as it is enclosed in the JSON Object called "event"
     -d '{"event": {"key1": "value1", "key2": "value2"}}'
```

# Data Structure in MongoDB

Here is an example of how a project with a single event will appear in MongoDB:

```json
{
  "_id": {
    "$oid": "650f49bcbad8d2c9a95ce18d"
  },
  "name": "test_project",
  "apikey": "abc123",
  "events": [
    {
      "timestamp": {
        "$date": "2023-09-23T20:26:28.573Z"
      },
      "Event": {
        "1": 2,
        "420": 69,
        "blah": "blerg",
        "bah bah black sheep": "Have you any wool?"
      }
    }
  ]
}
```
