import { HttpClient } from '@angular/common/http';
import { HttpService } from 'src/Utils/httpService';
import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'perseal-ui';
}
