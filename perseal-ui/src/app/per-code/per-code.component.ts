import { environment } from './../../environments/environment.prod';
import { ActivatedRoute } from '@angular/router';
import { HttpService } from 'src/Utils/httpService';
import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-per-code',
  templateUrl: './per-code.component.html',
  styleUrls: ['./per-code.component.css']
})
export class PerCodeComponent implements OnInit {

  code: string

  constructor(private server: HttpService, private route: ActivatedRoute) { }

  ngOnInit(): void {
    this.route.queryParams.subscribe(params =>
      this.code = params['code']
    )
    this.sendCode()
  }

  sendCode() {

    this.server.sendCode(this.code).subscribe(data => {

      }, error => {
        console.log(error);
      });

    window.location.href = environment.settings.host + '/insertPassword';
}
}
