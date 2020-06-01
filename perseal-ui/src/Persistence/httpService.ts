
import { Injectable, NgZone } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../environments/environment.prod';

@Injectable({
  providedIn: 'root'
})
export class  HttpService {

  constructor(private http: HttpClient) { }


  url = environment.settings.goservice;

  sendPassword(password: string,){
    return this.http.post(this.url + 'insertPassword', password);
  }

  requestDataCloudFiles(){
    return this.http.get(this.url + 'fetchCloudFiles');
  }

  requestRedirect(sessionId: string){
    return this.http.post(this.url + 'requestRedirect', sessionId);
  }

  sendCode(code: string){
    return this.http.post(this.url + 'code', code);
  }

  perStore(msToken: string){
    return this.http.post(this.url + 'store', msToken);
  }

  perLoad(msToken: string){
    return this.http.post(this.url + 'load', msToken);
  }

  noFilesStore(bool: boolean){
    return this.http.post(this.url + 'toStore', bool);
  }

  getSessionId(token: string){
    return this.http.post(this.url + 'getSessionId', token);
  }
}
