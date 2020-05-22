import { HomeComponent } from './home/home.component';
import { PerLoadComponent } from './per-load/per-load.component';
import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { GetPasswordComponent } from './get-password/get-password.component';


const routes: Routes = [
  {path: 'insertPassword', component: GetPasswordComponent, pathMatch: 'full'},
  {path: 'insertDataStoreFileName', component: PerLoadComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
