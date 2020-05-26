import { environment } from './../../environments/environment.prod';
import { HttpService } from 'src/Utils/httpService';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, Inject } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

@Component({
  selector: 'app-per-load',
  templateUrl: './per-load.component.html',
  styleUrls: ['./per-load.component.css']
})
export class PerLoadComponent implements OnInit {

  dataStoreFile: string;
  files: any;
  method: string;
  redirectUrl: HttpErrorResponse ;

  constructor(private server: HttpService, private route: ActivatedRoute) {

   }

  ngOnInit(): void {
    this.route.queryParams.subscribe(params =>
        this.method = params['method']
      )
    if(this.method === "load"){
    this.server.requestDataCloudFiles().subscribe(files =>
      this.files = files
    );
    }
  }

  sendDataStoreFile(password: string) {
    console.log(this.dataStoreFile)

    this.server.sendDataStoreFile(this.dataStoreFile, this.method).subscribe(data => {

      console.log("boa")
      window.location.href = environment.settings.host+ '/insertPassword';
      }, error => {

        console.log("mau")
        this.redirectUrl = error
        if(this.redirectUrl != null) {

          console.log(this.redirectUrl.error)
          console.log(this.redirectUrl.status)
          window.location.href=this.redirectUrl.error
          }
      });
    }
  storeFile(){
    this.method = "store"
  }
}
