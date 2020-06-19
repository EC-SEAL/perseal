# Persistence module for SEAL
## Microservice running on HTTP port 8082
Endpoint routes referenced in [perseal/routes.go](perseal/routes.go)

## To Do
### Mobile Storage
```
Affects: UC 1.07 and 8.03
```

## Usage

```bash
# Docker
docker-compose build
docker-compose up
```

#Locally (in the IDE)

```bash
set "Local" variable to true in "model" package
go run . (on package "perseal")
```

### How to Use
#### UC1.02 dashboard app local datastore access
```bash
http://vm.project-seal.eu:8082/per/load/{sessionId}?cipherPassword={pwd}
form-data: dataStore={dataStore}

dataStore is decrypted and stored as sessionVariable
on success returns: sessionId
```

#### UC1.03 dashboard app cloud datastore access
```bash
http://vm.project-seal.eu:8082/per/load
form-data: msToken={msToken} (with sessionVariable PDS=googleDrive or oneDrive)

dataStore is decrypted and stored as sessionVariable
on success returns: dataStore
```

#### UC1.05 dashboard first access local datastore access
```bash
http://vm.project-seal.eu:8082/per/load/{sessionId}?cipherPassword={pwd}
form-data: dataStore={dataStore}

dataStore is decrypted and stored as sessionVariable
on success returns: sessionId
```

#### UC1.06 dashboard first access cloud datastore access
```bash
http://vm.project-seal.eu:8082/per/load
form-data: msToken={msToken} (with sessionVariable PDS=googleDrive or oneDrive)

dataStore is decrypted and stored as sessionVariable
on success returns: dataStore
```

#### UC1.07 access web dashboard app with mobile datastore
```bash
http://vm.project-seal.eu:8082/per/load
form-data: msToken={msToken} (with sessionVariable PDS=Mobile)

dataStore from Dashboard Mobile App is decrypted and stored as sessionVariable
on success returns: sessionId
```

#### UC2.01 configures existing local mobile PDS in settings
```bash
http://vm.project-seal.eu:8082/per/load/{sessionId}?cipherPassword={pwd}
form-data: dataStore={dataStore}

dataStore is decrypted and stored as sessionVariable
on success returns: sessionId
```

#### UC2.02 first save to local browser datastore
```bash
http://vm.project-seal.eu:8082/per/load/{sessionId}?cipherPassword={pwd}
form-data: dataStore={dataStore}

dataStore is decrypted and stored as sessionVariable
on success returns: sessionId
```

#### UC2.04 dashboard setting configure and load cloud datastore
```bash
http://vm.project-seal.eu:8082/per/load
form-data: msToken={msToken} (with sessionVariable PDS=googleDrive or oneDrive)

dataStore is decrypted and stored as sessionVariable
on success returns: dataStore
```

#### UC2.05 dashboard setting configure and create cloud datastore
```bash
http://vm.project-seal.eu:8082/per/store
form-data: msToken={msToken} (with sessionVariable PDS=googleDrive or oneDrive)

dataStore is encrypted and stored on cloud storage
```

#### UC4.01 copy localMobile dataStore to localBrowser datastore
```bash
http://vm.project-seal.eu:8082/per/store/{sessionId}?cipherPassword={pwd}

dataStore is encrypted
on success returns: encrypted dataStore
```

#### UC8.03 SP Attribute Retrieval from Mobile PDS
```bash
http://vm.project-seal.eu:8082/per/load
form-data: msToken={msToken} (with sessionVariable PDS=Mobile)

dataStore from Dashboard Mobile App is decrypted and stored as sessionVariable
on success returns: sessionId
```

#### UC8.06 SP Attribute Retrieval from Browser PDS
```bash
http://vm.project-seal.eu:8082/per/load
form-data: msToken={msToken} (with sessionVariable PDS=Browser)

dataStore from Browser is decrypted and stored as sessionVariable
on success returns: msToken generated in Persistence Module
```

#### UC8.07 SP Attribute Retrieval from Cloud PDS
```bash
http://vm.project-seal.eu:8082/per/load
form-data: msToken={msToken} (with sessionVariable PDS=googleDrive or oneDrive)

dataStore from Cloud is decrypted and stored as sessionVariable
on success returns: msToken generated in Persistence Module
```