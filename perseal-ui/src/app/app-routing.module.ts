import { PerComponent } from './per/per.component';
import { LoadComponent } from './load/load.component';
import { StoreComponent } from './store/store.component';
import { PerCodeComponent } from './per-code/per-code.component';
import { PreConfigComponent } from './pre-config/pre-config.component';
import { NgModule, Component } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { GetPasswordComponent } from './get-password/get-password.component';


const routes: Routes = [
  {path: 'insertPassword', component: GetPasswordComponent, pathMatch: 'full'},
  {path: 'preConfig', component: PreConfigComponent},
  {path: 'code', component: PerCodeComponent},
  {
    path: 'per/:method',
    component: PerComponent,
    children : [
      {path: 'store', component: StoreComponent},
      {path: 'load', component: LoadComponent},
    ]
}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
