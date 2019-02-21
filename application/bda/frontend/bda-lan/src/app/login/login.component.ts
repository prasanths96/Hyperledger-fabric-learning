import { Component, OnInit, OnDestroy } from '@angular/core';
import { FormControl, Validators, NgForm } from '@angular/forms';
import { AuthService } from '../services/auth.service';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material';
import { Subscription } from 'rxjs';


@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {

  user = new FormControl('', [Validators.required]);
  userpw = new FormControl('', [Validators.required]);
  loginFail = false;
  authStatusListener = this.authService.getAuthStatusListener();
  isLoading = false;
  private userIsAuthenticated = false;
  private authListenerSubs: Subscription;

  constructor(private authService: AuthService,
              private router: Router,
              private snackBar: MatSnackBar) { }

  ngOnInit() {
    this.userIsAuthenticated = this.authService.getIsAuth();
    this.authListenerSubs = this.authService
      .getAuthStatusListener()
      .subscribe(isAuthenticated => {
        this.userIsAuthenticated = isAuthenticated;
      });
    if(this.userIsAuthenticated){
      this.router.navigate(["/main"]);
      this.snackBar.open("Already logged in.", "OK", {
        duration: 2000,
      });
    }
    this.loginFail = false;
  }

  ngOnDestroy() {
    this.authListenerSubs.unsubscribe();
  }

  onLogin(form: NgForm) {
    if (form.invalid) {
      return;
    }
    console.log(form.value.user);
    this.authService.login(form.value.user, form.value.userpw);
    form.resetForm('');
    this.isLoading = true;
    this.authStatusListener.subscribe( isAuth => {
      this.isLoading = false;
      if(!isAuth){
        this.loginFail = true;
      } else {
        this.loginFail = false;
      }  
    });
  }

}
