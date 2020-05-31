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
  link: any
  ngOnInit(): void {
    this.route.queryParams.subscribe(params =>
      this.token = params['token']
    )
    this.openURL();
    setTimeout(() =>
    {
     window.open( environment.settings.host + '/preConfig?method=store');

    },
    1750);

  }

  async openURL(){
    this.server.perStore(this.token).subscribe(link =>{

      this.link = link;
      console.log("made post")
      window.location.href = this.link

    }, error => {
      console.log(error)
    })
  }

}
