# Persistence module for SEAL
* Microservice running on **HTTPS port 8082**
* Endpoint routes referenced in [perseal/routes.go](perseal/routes.go)

## To Do
* Integrate back channel requests with the mobile application;
* When PDS is Mobile, then validate if the user is on a mobile device or desktop to either show QRcode or redirect to custom URL;
* Integration with Dashboard: should the user have to create a PDS everytime he accesses the dashboard operations?;
* Test the cloud (GoogleDrive | OneDrive) token expiration;

## Usage

### Docker
on the *perseal* folder, run  `docker-compose build && docker-compose up`

### Locally (for development)
set the "Test" variable to true in *model/object.go* file 
on the *perseal* folder, run `go get ./... && go run .`

## Behaviour and Functionallities (for other microservices to use)

### Front-Channel Store
* **URL: GET /per/store{?msToken}**

1. Validates the token using the SM and extracts its contents from the "Additional Data" field;
2. Checks if a ciphered password exists in the query. If so, then it's considered a back-channel request;
3. Gets session data using the SM from the sessionId in the token;
4. Builds the DTO with that data;
5. Shows the respective HTML page depending on the the PDS location:
    5.1.  Mobile: generates a msToken, using the SM, containing the method "store" and the sessionId and sends it to the **/per/QRcode** endpoint in order to show it in a QRcode for the mobile application to read. It also updates the session variable "finishedPersealBackChannel" to the value "not finished", to control when the QRcode should disappear after the user performs the next back-channel load request;
    5.2.  Browser(Local): shows form to insert the file name and insert the password to encrypt the current data in session;
    5.3.  GoogleDrive and OneDrive: generates the **redirect_url** and redirects the user to that URL for the it to select his or her Google or Microsoft account. After the user selects the account, sends a code to the persistence endpoint /per/code.

### Front-Channel Load
* **URL: GET /per/load{?msToken}**

1. Validates the token using the SM and extracts its contents from the "Additional Data" field;
2. Checks if a ciphered password exists in the query. If so, then it's considered a back-channel request;
3. Gets session data using the SM from the sessionId in the token;
4. Builds the DTO with that data;
5. Shows the respective HTML page depending on the the PDS location:
    5.1.  Mobile: generates a msToken, using the SM, containing the method "laod" and the sessionId and shows it in a QRcode for the mobile application to read. It also updates the session variable "finishedPersealBackChannel" to the value "not finished", to control when the QRcode should disappear after the user performs the next back-channel load request;
    5.2.  Browser(Local): shows a page in which the user can select either to load a current PDS file OR an option similiar to the "store" funcionallity, where the user can insert the file name and insert the password to encrypt the current data in session. In the "Load Current File" case, the user selects the ".seal" file from his machine and, afterwards, inserts the password to decrypt the contents of that file;
    5.3.  GoogleDrive and OneDrive: generates the **redirect_url** and redirects the user to that URL for the it to select his or her Google or Microsoft account. After the user selects the account, sends a code to the persistence endpoint /per/code.
    
### Back-Channel Store
* **URL: GET /per/store{?msToken}{&cipherPassword}**

1. Validates the token using the SM and extracts its contents from the "Additional Data" field;
2. Checks if a ciphered password exists in the query;
3. Gets session data using the SM from the sessionId in the token;
4. Builds the DTO with that data;
5. Fetches the session data on the "AdditionalData" field (by using the **/sm/new/search** endpoint), encrypts it using AES and signs it using a RSA private key;
6. Returns the encrypted dataStore in the Response Body.

### Back-Channel Load
* **URL: POST/per/load{?msToken}{&cipherPassword}**
* RequestBody: encrypted DataStore

1. Validates the token using the SM and extracts its contents from the "Additional Data" field;
2. Gets session data using the SM from the sessionId in the token;
3. Builds the DTO with that data;
4. Validates the signature of the dataStore in the Request Body;
5. Decrypts the contents of the dataStore;
6. Clears the current session data by using "sm/new/startSession";
7. Adds the values of the currently decrypted dataStore using "sm/new/add";
8. Returns the decrypted dataStore in the Response Body.

## Behaviour and Functionallities (for internal perseal uses)

### Code
* **URL: GET /per/code{?state}{&code}*

1. Gets session data using the SM from the sessionId in the query parameter;
2. Builds the DTO with that data;
3. Updates the respective session variable ("GoogleDriveToken" or "OneDriveToken") with the token generated from the "code" variable
4. Shows the respective HTML page depending on the the PDS location:
    4.1.  Store:  shows form to insert the file name and insert the password to encrypt the current data in session;
    4.2.  Load:  gets SEAL files from the user's account storage. if no files are found, presents warning to the user and shows form to insert the file name and insert the password to encrypt the current data in session. Otherwise, shows file list to the user for it to choose one and insert the password to encrypt the file's contents.
 
### Generate QRCode
* **URL: GET /per/QRcode{?state}{&code}*

1. Validates the token using the SM and extracts its contents from the "Additional Data" field;
2. Gets session data using the SM from the sessionId in the token;
3. Builds the DTO with that data;
4. Generate the msToken, using the msToken, for the mobile application to read and shows it to the user as a QRcode.

### DataStore Handling: Store
* **URL: POST/per/insertPassword/store{?msToken}*

1. Validates the token using the SM and extracts its contents from the "Additional Data" field;
2. Checks if a ciphered password exists in the query. If so, then it's considered a back-channel request;
3. Gets session data using the SM from the sessionId in the token;
4. Builds the DTO with that data;
5. Encrypts and compiles the information according to the PDS location:
    5.1. Browser: fetches the session data on the "AdditionalData" field (by using the **/sm/new/search** endpoint), encrypts it using AES and signs it using a RSA private key. Downloads the file by using **/per/save** endpoint;
    5.2. GoogleDrive|OneDrive: checks if the tokens haven't expired and, then, fetches the session data on the "AdditionalData" field (by using the **/sm/new/search** endpoint), encrypts it using AES and signs it using a RSA private key. Uploads the file to the cloud storage;
6. Builds an "OK" msToken and sends it to the ClientCallbackAddr session variable, by using the **/per/pollcca** endpoint.
**NOTE: if the user is redirected to this endpoint after a GET /per/load request (e.g. after the user selects Load "Browser" PDS but chooses to create a new file instead of loading an existent one), the persistence module will store the file as described before and immediately load the information in the session's data (which most of the times will result in an empty dataStore)**

### DataStore Handling: Load
* **URL: POST/per/insertPassword/load{?msToken}*

1. Validates the token using the SM and extracts its contents from the "Additional Data" field;
2. Checks if a ciphered password exists in the query. If so, then it's considered a back-channel request;
3. Gets session data using the SM from the sessionId in the token;
4. Builds the DTO with that data;
5. Fetches the PDS file from the respective source and validates if it contains a valid structure;
6. Validates the signature of the dataStore;
7. Decrypts the contents of the dataStore;
8. Clears the current session data by using "sm/new/startSession";
9. Adds the values of the currently decrypted dataStore using "sm/new/add";
10.  Builds an "OK" msToken and sends it to the ClientCallbackAddr session variable, by using the **/per/pollcca** endpoint.

### Poll MSToken to the ClientCallbackAddr (CCA)
* **URL: GET/per/pollcca{?msToken}{&tokenInfo}*

1. Validates the **msToken** using the SM and extracts its contents from the "Additional Data" field;
2. If valid, then send the **tokenInfo** as a POST request to the ClientCallbackAddr session variable;



