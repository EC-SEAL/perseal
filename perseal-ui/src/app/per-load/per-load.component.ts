import { HttpService } from 'src/Utils/httpService';
import { HttpClient } from '@angular/common/http';
import { Component, OnInit, Inject } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

@Component({
  selector: 'app-per-load',
  templateUrl: './per-load.component.html',
  styleUrls: ['./per-load.component.css']
})
export class PerLoadComponent implements OnInit {

  dataStoreFile: string;
  files: any;
  _sessionToken: string

  constructor(private server: HttpService, private route: ActivatedRoute) {

   }

  ngOnInit(): void {
    this.route.queryParams.subscribe(params =>
        this._sessionToken = params['sessionToken']
      )
    this.server.requestDataCloudFiles(this._sessionToken).subscribe(files =>
      this.files = files
    );
  }

  sendDataStoreFile(password: string) {
    this.server.sendDataStoreFile(this.dataStoreFile).subscribe(data => {

      }, error => {
        console.log(error);
      });
    window.location.href = 'http://localhost:4200/insertPassword';
}
}
