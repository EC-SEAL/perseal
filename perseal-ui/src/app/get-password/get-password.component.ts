import { ActivatedRoute } from '@angular/router';
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
  method: string;


  constructor(private server: HttpService,private route: ActivatedRoute) { }

   ngOnInit() {
    this.route.queryParams.subscribe(params =>
      this.method = params['method']
    )
  }

  sendPassword(password: string) {
    this.server.sendPassword(this.password, this.method).subscribe((data: HttpErrorResponse) => {

      }, error => {
      });

    }
}
