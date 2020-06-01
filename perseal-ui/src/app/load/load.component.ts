import { ActivatedRoute } from '@angular/router';
import { HttpService } from 'src/Persistence/httpService';
import { Component, OnInit } from '@angular/core';
import { environment } from './../../environments/environment.prod';

@Component({
  selector: 'app-load',
  templateUrl: './load.component.html',
  styleUrls: ['./load.component.css']
})
export class LoadComponent implements OnInit {

  constructor(private server: HttpService, private route: ActivatedRoute) { }

 token: string
  link: any
  toStore: string
  sessionId: any
  error: boolean

  ngOnInit(): void {
    this.error = false
    this.route.queryParams.subscribe(params =>
      this.token = params['token']
    )
    this.server.getSessionId(this.token).subscribe(sessionId => {
      this.sessionId = sessionId;
      this.openURL();
      setTimeout(() =>
      {
       window.open( environment.settings.host + '/preConfig?method=load');

      },
      1750);

    },  error =>{
      console.log('falhou')
      this.error = true
    }
    )
  }

  async openURL(){
    this.server.perLoad(this.sessionId).subscribe(link =>{

      this.link = link;
      console.log("made post")
      window.location.href = this.link

    }, error => {
      console.log(error)
    })
  }

}
