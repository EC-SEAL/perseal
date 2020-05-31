import { HttpService } from 'src/Persistence/httpService';
import { Component, OnInit, Input } from '@angular/core';
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

  constructor(private server: HttpService) { }

   ngOnInit() {
    this.toStore="load";
    console.log(this.toStore)
    this.server.requestDataCloudFiles().subscribe(files => {
      this.files = files;
      console.log(this.files);
      if (this.files != null) {

      if (Object.keys(this.files).length !== 0) {
        this.server.noFilesStore(false).subscribe((data: HttpErrorResponse) => {

        }, error => {
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


  storeFile(){
    this.toStore = "store";
    this.server.noFilesStore(true).subscribe((data: HttpErrorResponse) => {

    }, error => {
    });
  }
}
