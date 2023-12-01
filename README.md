# Note-App-Microservices
> A collection of microservices that make up a note app backend
---

## Note API Documentation

- Create a document via curl: `curl -i -X Post -d '{BODY}' URL/create`
- Read a document via curl: `curl -i URL/find/{user}/{title}` or `curl -i URL/find/many/{user}/{title}`
- Update a document via curl: `curl -i -X Put -d  '{BODY} URL/{user}/{title}'`
- Delete a document via curl: `curl -i -X Delete -d URL/delete/{user}/{title}`

## User API Documentation

- The user API is tied directly to a frontend html form, so I'm not sure it can be called by a curl command efficiently
- Will probably refactor this to work with JSON body requests
