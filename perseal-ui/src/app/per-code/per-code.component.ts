import { StoreComponent } from './../store/store.component';
import { environment } from './../../environments/environment.prod';
import { ActivatedRoute } from '@angular/router';
import { HttpService } from 'src/Persistence/httpService';
import { Component, Output, EventEmitter, OnInit } from '@angular/core';

@Component({
  selector: 'app-per-code',
  templateUrl: './per-code.component.html',
  styleUrls: ['./per-code.component.css']
})
export class PerCodeComponent implements OnInit {

  code: string

  constructor(private server: HttpService, private route: ActivatedRoute) { }

  ngOnInit(): void {
    this.route.queryParams.subscribe(params =>{
      this.code = params['code']
      console.log(this.code)
    })
    console.log('Recieved Code');
    this.sendCode()
  }


  sendCode() {

    this.server.sendCode(this.code).subscribe(data => {
      window.location.href = environment.settings.host + "/insertPassword"
      }, error => {
        console.log(error);
      });
}
}
