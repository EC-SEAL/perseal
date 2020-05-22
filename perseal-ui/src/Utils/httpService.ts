
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
    return this.http.post(this.url + 'insertPassword?method=' + method, password);
  }

  requestDataCloudFiles(){
    return this.http.get(this.url + 'fetchCloudFiles');
  }

  sendDataStoreFile(filename: string, method: string){
    return this.http.post(this.url + 'insertDataStoreFilename?method=' + method, filename);
  }

  sendCode(code: string){
    return this.http.post(this.url + 'code', code);
  }


}
