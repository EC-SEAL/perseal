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

  ngOnInit(): void {
    this.route.queryParams.subscribe(params =>
      this.token = params['token']
    )
    this.openURL();
    setTimeout(() =>
    {
     window.open( environment.settings.host + '/preConfig?method=load');

    },
    1750);

  }

  async openURL(){
    this.server.perLoad(this.token).subscribe(link =>{

      console.log("made post")

    }, error => {
      console.log(error)
    })
  }

}
