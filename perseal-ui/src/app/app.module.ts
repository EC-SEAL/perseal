import { HttpService } from 'src/Persistence/httpService';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { GetPasswordComponent } from './get-password/get-password.component';

import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { PreConfigComponent } from './pre-config/pre-config.component';
import { PerCodeComponent } from './per-code/per-code.component';
import { LoadComponent } from './load/load.component';
import { StoreComponent } from './store/store.component';

@NgModule({
  declarations: [
    AppComponent,
    GetPasswordComponent,
    PreConfigComponent,
    PerCodeComponent,
    LoadComponent,
    StoreComponent,

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
