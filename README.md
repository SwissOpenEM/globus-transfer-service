# Globus Transfer Service

### Summary

This service allows for requesting [Globus](https://www.globus.org) transfers using the [Scicat](https://scicatproject.github.io) token. It also keeps track of ongoing transfers, and their state can be polled from this service.

It relies on a single [service account](https://docs.globus.org/guides/recipes/automate-with-service-account/) to request and track transfers, which can be set using environment variables.

It also assumes that there's a set of possible destinations and sources. Whether it's possible to ingest from the source to the given destination is defined by the Scicat user's current set of groups. The group associations are given to this service as a config. 

### Configuration

 - `facilityCollectionIDs` - a map of facility names (identifiers) to their collection id's
 - `globusScopes` - the scopes to use for the client connection. Access is required to transfer api and specific collections
 - `port` - the port at which the server should run

An example is provided at `internal/config/example-config.yaml`

### Environment variables

 - `GLOBUS_CLIENT_ID` - the client id for the service account (2-legged OAUTH, trusted client model)
 - `GLOBUS_CLIENT_SECRET` - the client secret for the service account (2-legged OAUTH, trusted client model)
