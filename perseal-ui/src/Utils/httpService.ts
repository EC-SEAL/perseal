
import { Injectable, NgZone } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})

export class  HttpService {

  constructor(private http: HttpClient) { }


  url = 'http://localhost:8082/per/';

  sendPassword(password: string, method: string){
    return this.http.post(this.url + 'insertPassword?method='+method, password);
  }

  requestDataCloudFiles(sessionId: string){
    return this.http.get(this.url + 'fetchCloudFiles?sessionToken=' + sessionId);
  }

  sendDataStoreFile(filename: string){
    return this.http.post(this.url + 'insertDataStoreFilename', filename);
  }

}
