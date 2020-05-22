import { Injectable } from '@angular/core';
import { HttpService } from 'src/Utils/httpService';

@Injectable({
  providedIn: 'root'
})
export class GoPersistence {

  constructor(private http: HttpService) { }

  output: JSON;
  obj: any;


  sendDataStore(filename: string): Promise<any>{
    return new Promise<any> ((resolve, reject) => {
      this.http.sendDataStoreFile(filename, "stroe").subscribe(data => {
        resolve(data as any);
      });
    });
  }

}
