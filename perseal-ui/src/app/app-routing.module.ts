import { PerCodeComponent } from './per-code/per-code.component';
import { PerLoadComponent } from './per-load/per-load.component';
import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { GetPasswordComponent } from './get-password/get-password.component';


const routes: Routes = [
  {path: 'insertPassword', component: GetPasswordComponent, pathMatch: 'full'},
  {path: 'insertDataStoreFileName', component: PerLoadComponent},
  {path: 'code', component: PerCodeComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
