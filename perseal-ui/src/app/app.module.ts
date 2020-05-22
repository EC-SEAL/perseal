import { HttpClient, HttpClientModule } from '@angular/common/http';
import { GetPasswordComponent } from './get-password/get-password.component';

import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { PerLoadComponent } from './per-load/per-load.component';
import { PerCodeComponent } from './per-code/per-code.component';

@NgModule({
  declarations: [
    AppComponent,
    GetPasswordComponent,
    PerLoadComponent,
    PerCodeComponent,

  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    FormsModule,
    HttpClientModule
  ],
  providers: [HttpClient],
  bootstrap: [AppComponent],
})
export class AppModule { }
