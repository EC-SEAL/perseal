import { HttpService } from 'src/Persistence/httpService';
import { ActivatedRoute } from '@angular/router';
import { Component, OnInit } from '@angular/core';
import { environment } from 'src/environments/environment.prod';

@Component({
  selector: 'app-per',
  templateUrl: './per.component.html',
  styleUrls: ['./per.component.css']
})
export class PerComponent implements OnInit {


  id: string
  token: string
  error: boolean
  sessionId: any
  link : any

  constructor(
    private route: ActivatedRoute,
    private server: HttpService
  ) {}

  ngOnInit(): void {

    this.server.reset().subscribe(link =>{
      console.log("reset")
    }, error => {
    })

    this.error = false
    this.id = this.route.snapshot.paramMap.get('method');
    this.route.queryParams.subscribe(params =>
      this.token = params['token']
    );
    this.route.queryParams.subscribe(params =>
      this.token = params['token']
    )
    this.server.getSessionId(this.token).subscribe(sessionId => {
      this.sessionId = sessionId;
      if (this.id == "store"){
      this.store();
      setTimeout(() =>
      {
       window.open( environment.settings.host + '/preConfig?method=store');

      },
      1750);

    }
      if (this.id == "load"){
      this.load();
      setTimeout(() =>
      {
       window.open( environment.settings.host + '/preConfig?method=load');

      },
      1750);

    }
    },  error =>{
      console.log('falhou')
      this.error = true
    }
    )
  }

  async store(){
    this.server.perStore(this.sessionId).subscribe(link =>{

      this.link = link;
      console.log("made post")
      window.location.href = this.link

    }, error => {
      console.log(error)
    })
  }

  async load(){
    this.server.perLoad(this.sessionId).subscribe(link =>{

      this.link = link;
      console.log("made post")
      window.location.href = this.link

    }, error => {
      console.log(error.url)
    })
  }

}
