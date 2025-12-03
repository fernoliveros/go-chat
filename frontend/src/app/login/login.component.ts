import { Component } from '@angular/core';
import { AuthService } from '../auth.service';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { Observable, Subscription } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { Router } from '@angular/router';

@Component({
    selector: 'app-login',
    templateUrl: './login.component.html',
    styleUrls: ['./login.component.scss'],
    standalone: false
})
export class LoginComponent {
  loginForm = new FormGroup({
    username: new FormControl('', Validators.required),
    password: new FormControl('', Validators.required),
  });

  private loginSub: Subscription = new Subscription();

  constructor(private authService: AuthService, private router: Router) {}

  login() {
    this.loginSub = this.authService.login(this.loginForm.value).subscribe({
      next: (value) => {
        console.log('successful login', value);
        localStorage.setItem('authenticated', 'true');
        localStorage.setItem("username", this.loginForm.value.username ?? "Guest")
        this.router.navigate(['/']);
      },
      error: (err) => {
        console.log('Login Error', err);
        this.loginForm.get('password')?.reset();
        localStorage.setItem('authenticated', 'false');
      },
      complete: () => {
        console.log('Login observable completed');
      },
    });
  }

  ngOnDestroy(): void {
    console.log('ngOnDestroy!!!');
    if (this.loginSub) {
      this.loginSub.unsubscribe();
    }
  }
}
