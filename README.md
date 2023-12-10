# Note-App-Microservices
> A collection of microservices that make up a note app backend
---

## Note API Documentation

- Create a document via curl: `curl -i -X Post -d '{BODY}' URL/create`
- Read a document via curl: `curl -i URL/find/{user}/{title}` or `curl -i URL/find/many/{user}/{title}`
- Update a document via curl: `curl -i -X Put -d  '{BODY} URL/{user}/{title}'`
- Delete a document via curl: `curl -i -X Delete -d URL/delete/{user}/{title}`

### Note Schema
```
ID          primitive.ObjectID `bson:"_id"`
Title       string             `bson:"title"`
Description string             `bson:"description"`
User        string             `bson:"user"`
Date        time.Time          `bson:"date"`
```

## User API Documentation

- [ ] Integrate Json body requests
- [ ] Complete Authentication for client

### User Schema
```
username    string `bson:"username"`
password    string `bson:"username"`
email       string `bson:"email"`
DateOfBirth string `bson:"date of birth"`
```
