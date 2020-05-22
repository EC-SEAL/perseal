import { HttpService } from 'src/Utils/httpService';
import { Component, OnInit } from '@angular/core';
import { HttpResponse, HttpErrorResponse } from '@angular/common/http';

@Component({
    selector: 'app-get-password',
    templateUrl: './get-password.component.html',
    styleUrls: ['./get-password.component.css']
})
export class GetPasswordComponent implements OnInit {

  password: string;
  redirectUrl: HttpErrorResponse;


  constructor(private server: HttpService) { }

   ngOnInit() {
  }

  sendPassword(password: string) {
    this.server.sendPassword(this.password, "store").subscribe((data: HttpErrorResponse) => {

      }, error => {
        this.redirectUrl = error
        if(this.redirectUrl != null) {

          console.log(this.redirectUrl.error)
          console.log(this.redirectUrl.status)
          window.location.href=this.redirectUrl.error
          }
      });

    }
}
