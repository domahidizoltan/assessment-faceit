# Faceit - User management service

author: Zoltán Domahidi  
date: November 28, 2022  

<br/>

### TASK

To write a small microservice to manage access to Users, the service should be implemented in Go - as this is primary language that we use at FACEIT for backend development.

### REQUIREMENTS

A user must be stored using the following schema:
```
{
    "id": "d2a7924e-765f-4949-bc4c-219c956d0f8b",
    "first_name": "Alice",
    "last_name": "Bob", 
    "nickname": "AB123", 
    "password": "supersecurepassword",
    "email": "alice@bob.com",
    "country": "UK",
    "created_at": "2019-10-12T07:20:50.52Z",
    "updated_at": "2019-10-12T07:20:50.52Z"
}
```

The service must allow you to:  
    • Add a new User  
    • Modify an existing User  
    • Remove a User  
    • Return a paginated list of Users, allowing for filtering by certain criteria (e.g. all Users with the country "UK")  

The service must:  
    • Provide an HTTP or gRPC API  
    • Use a sensible storage mechanism for the Users  
    • Have the ability to notify other interested services of changes to User entities  
    • Have meaningful logs  
    • Be well documented  
    • Have a health check  

The service must NOT:  
    • Provide login or authentication functionality  

It is up to you what technologies and patterns you use to implement these features, but you will be assessed on these choices and we expect you to be confident in explaining your choice. We encourage the use of local alternatives or stubs (for instance a database containerised and linked to your service through docker-compose). 


#### Notes

Remember that we want to test your understanding of these concepts, not how well you write boilerplate code. If your solution is becoming overly complex, simply explain what would have been implemented and prepare for follow-up questions in the technical interview.

Please also provide a README.md that contains:  
    • Instructions to start the application on localhost (dockerised applications are preferred)  
    • An explanation of the choices taken and assumptions made during development  
    • Possible extensions or improvements to the service (focusing on scalability and deployment to production)  

We expect to be able to run the tests, build the application and run it locally.


#### WHAT YOU WILL BE SCORED ON

Coding Skills:  
    • Is your code respecting fields and access modifiers?  
    • Is your code respecting single responsibility principles?  

Application Structure:  
    • Have you applied the correct division of the layers?  
    • Do you have the correct dependencies between layers?  

Framework Usage:  
    • Have you applied the correct usage of framework features?  

REST endpoints:  
    • Is your URL structure correct?  
    • Have you used HTTP verbs?  

Asynchronous communication:  
    • Is it asynchronous?  

Testing:  
    • Are you happy with your test coverage?  

---

### Running the service 

The service could be configured by environment variables and by default it will use connections on localhost (see the commented out `docker-compose.yaml` for the env vars)

With the help of `make` commands it is possible to build and run the service: 
- `make install` will install the required Go tools
- `make generate` will re-generate the generated files and updates the workspace
- `make compose-up` starts the dockerized environment
  - optionally now you can run the integration tests with `make test`
  - now you can access the middlewares and tools:
    - RabbitMQ on `localhost:15672` (guest:guest)
    - Adminer (UI to the DB) on `localhost:9000` (admin:pass)
    - Swagger Editor on `localhost:80`
- `make run` starts the service and could be accessed on `localhost:8000`
    - eventually you can bind a queue to `events.user` exchange with `#` binding to see the published user events

> The service could run dockerized from the Docker Compose environment. For this you must uncomment the `userservice` section in `docker-compose.yaml` and change the server host in Swagger Editor to the container IP. You can get the container IP with this command:
> ```
> docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' faceit_userservice_1
> ```
>  

In case `make` is not installed the commands could be run from the `Makefile` instead. 

<br/>

### Explanation of choices and assumptions

- The REST API is defined in the `api/users.yaml` OpenAPI file what is also used to generate the routing and types of the handler layer
- The `docker-compose.yaml` also has a Swagger Editor (`localhost:80`) and an Adminer (`localhost:9000`) to check and test the REST API and the database.
- The service is using Postgres database (I was more familiar with Postgres and at this level I don't think that different database would make huge difference for this data model)
- The service contains the public facing layers in the `pkg` directory (i.e. server definition, DTOs and handlers) and the internal layers in the `internal` directory (i.e. service and repository/DAL)
- Makery is used to generate mocks from the interfaces
- Password is not stored in the main `User` data model because it is very senseitive and should not be exposed in any way so it is handled separately and it couldn't be read, only written and changed
- Errors are handled, logged and transformed to REST response in `pkg/user/error_handler.go`
- Events are published over RabbitMQ so the consumers could receive events when they become online. The service is responsible only to create the topic exchange to broadcast the user events. The consumers are responsible for creating the queues. This way the exchange hides the queue topology and it's changes from the producer (the service).
- The tests follow the testing pyramid principles (layer behaviour is tested with unit tests, IO related operations (Http request, database operation) are covered with integration tests, and there are some API tests to see that the layers and frameworks are working together)
- The health endpoint could be found at `/health` and it is undocumented

<br/>

### Possible extensions or improvements

- Password is encrypted with SHA256 at the moment but it could have a proper encryption (what could also decrypt the password if that's required)
- Log details and stack traces could be improved
- Event publishing could be asynchronous so it won't block the call chain (this could be important on high load). It could be also more reliable by having any retry mechanism to increase it's fault tolerance
- Events could contain only the changed data (i.e. on update)
- Application configuration could be refactored to have in a central place using a proper config library (i.e. Viper)
- Server could have graceful shutdown to not interrupt ongoing requests and event publications when a shutdown signal was received
- Better organization of common (not strictly user related) constants, models and helpers
- Health check and RabbitMQ connection should be recover after an RMQ outage
- Data filtering (`GET /users`) is using only AND operator because of simplicity. OR operator or some more complex filtering could be made but probably that requires to use a `POST` operation with a payload which defines the filter operations.

---

### Outcome

I think I made everything what was required:
- the service can manage users
- it is well tested and documented
- it sends event notifications
- it has healthcheck
- it is using a dockerized environment
- it is listing users with pagination and filters
- plus it uses `PATCH` for updating users instead of `PUT`

The application was rejected with a single feedback: `they believe the code could be better structured`

I rather not comment on this :)