import { Component, OnInit } from '@angular/core';
import { environment } from './../../environments/environment.prod';
import { ActivatedRoute } from '@angular/router';
import { HttpService } from 'src/Persistence/httpService';

@Component({
  selector: 'app-store',
  templateUrl: './store.component.html',
  styleUrls: ['./store.component.css']
})
export class StoreComponent implements OnInit {

  constructor(private server: HttpService, private route: ActivatedRoute) { }

  token: string
  sessionId: any
  link: any
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
       window.open( environment.settings.host + '/preConfig?method=store');

      },
      1750);

    },  error =>{
      console.log('falhou')
      this.error = true
    }
    )

  }

  async openURL(){
    this.server.perStore(this.sessionId).subscribe(link =>{

      this.link = link;
      console.log("made post")
      window.location.href = this.link

    }, error => {
      console.log(error)
    })
  }

}
