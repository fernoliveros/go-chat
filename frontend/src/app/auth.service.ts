import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  apiUrl: string = environment.apiUrl

  constructor(private http: HttpClient) {}

  isAuthenticated(): boolean {
    const isAuthd = localStorage.getItem('authenticated');
    return isAuthd === 'true';
  }

  login(form: object): Observable<any> {
    const headers = new HttpHeaders({ 'Content-Type': 'application/json' });
    return this.http.post(this.apiUrl + '/login', form, { headers });
  }
}
