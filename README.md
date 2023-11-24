# load-balancer.go

## Motivation

This personal project serves as a practice to explore how a load balancer works. It contains the following:

- Load balancer service
- Backend service

### Load balancer service

Uses round robin algorithm to distribute the incoming requests. Few endpoints available:

- `GET /register` - Get list of registered server
- `POST /register?server={ADDRESS_OF_SERVER}` - Register a new backend service to the load balancer
- `DELETE /register?server={ADDRESS_OF_SERVER}` - Remove a backend service from the load balancer

Health check for each backend service is also available. Should one of the backend unhealthy, request will be diverted to the next available service.

Other request will be forwarded to the backend respectively.

### Backend service

A simple service that accepts any type of request and print out the request information.

This service has a specific `GET /health` endpoint implemented that is pinged by the load balancer to check the status of the service.

## Development

1. Run `make all` to compile the code and create executables for both load balancer and the backend service
2. Run `make run-lb` to start the load balancer. It will run on port 8080.
