import { environment } from '../../environments/environment.prod';
import { HttpService } from 'src/Persistence/httpService';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, Inject, HostListener } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

@Component({
  selector: 'app-pre-config',
  templateUrl: './pre-config.component.html',
  styleUrls: ['./pre-config.component.css']
})
export class PreConfigComponent implements OnInit {

  dataStoreFile: string;
  files: any;
  method: string;
  redirectUrl: HttpErrorResponse ;
  hasRedirect: any;
  sessionId: string;

  constructor(private server: HttpService, private route: ActivatedRoute) {

   }

  ngOnInit(): void {
    this.route.queryParams.subscribe(params =>
        this.method = params['method']
      );
    this.route.queryParams.subscribe(params =>
        this.sessionId = params['sessionId']
      );
    if(this.method === "load"){

      this.server.requestRedirect(this.sessionId).subscribe(hasRedirect =>{
        console.log("certo")
        this.hasRedirect = false;
        window.location.href = environment.settings.host + '/insertPassword';
      }, error =>{
        console.log('mau');
        this.hasRedirect = true
        this.redirectUrl = error;
        if(this.redirectUrl != null) {

          console.log(this.redirectUrl.error);
          console.log(this.redirectUrl.status);
          window.location.href=this.redirectUrl.error;
          }
      });

    } else  if(this.method === "store"){
      this.server.requestRedirect(this.sessionId).subscribe(hasRedirect =>{
        this.hasRedirect = false;
        window.location.href = environment.settings.host + '/insertPassword';
      }, error => {
        console.log('mau');
        this.hasRedirect = true
        this.redirectUrl = error;
        if(this.redirectUrl != null) {

          console.log(this.redirectUrl.error);
          console.log(this.redirectUrl.status);
          window.location.href=this.redirectUrl.error
          }
      });
    }
  }

}
