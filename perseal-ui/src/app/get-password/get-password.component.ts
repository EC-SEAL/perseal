import { ActivatedRoute } from '@angular/router';
import { HttpService } from 'src/Persistence/httpService';
import { Component, OnInit, Input, HostListener } from '@angular/core';
import { HttpResponse, HttpErrorResponse } from '@angular/common/http';

@Component({
    selector: 'app-get-password',
    templateUrl: './get-password.component.html',
    styleUrls: ['./get-password.component.css']
})
export class GetPasswordComponent implements OnInit {

  password: string;
  files: any
  toStore: string;
  sessionId: string
  link: any


  constructor(private server: HttpService, private route: ActivatedRoute) { }

   ngOnInit() {

    this.route.queryParams.subscribe(params =>
      this.sessionId = params['sessionId']
    );
    this.toStore="load";
    console.log(this.toStore)

    this.server.clientCallbackAddr().subscribe(link => {
      this.link = link;
      console.log(this.link);
    }, error => {
      console.log(error)
    });


    this.server.requestDataCloudFiles().subscribe(files => {
      this.files = files;
      console.log(this.files);
      if (this.files != null) {

      if (Object.keys(this.files).length !== 0) {
        this.server.noFilesStore(false).subscribe(link => {
          console.log(link)
        }, error => {
          console.log(error)
        });
      }
   }
    },error => {
    });

  }

  sendPassword(password: string) {
    this.server.sendPassword(password).subscribe((data: HttpErrorResponse) => {
        window.close()
      }, error => {
      });

    }

    goBack(){

      window.location.href = this.link
    }

  storeFile(){
    this.toStore = "store";
    this.server.noFilesStore(true).subscribe((data: HttpErrorResponse) => {

    }, error => {
    });
  }

  @HostListener('window:beforeunload', [ '$event' ])
  beforeUnloadHandler(event) {
    this.server.resetAndClose().subscribe(link => {

    }, error => {
    });
  }
}
