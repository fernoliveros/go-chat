import { Injectable, NgZone } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, Observer } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class MessageService {
  private apiUrl = 'http://localhost:8080';

  constructor(private http: HttpClient, private ngZone: NgZone) {}

  sendMessage(form: object): Observable<any> {
    const headers = new HttpHeaders({ 'Content-Type': 'application/json' });
    return this.http.post(this.apiUrl + '/send', form, { headers });
  }

  getMessagesSSE(): Observable<MessageEvent> {
    return new Observable((observer: Observer<MessageEvent>) => {
      const eventSource = new EventSource(this.apiUrl + '/messages');

      eventSource.onmessage = (event) => {
        this.ngZone.run(() => {
          observer.next(event);
        });
      };

      eventSource.onerror = (error) => {
        this.ngZone.run(() => {
          observer.error(error);
        });
      };

      return () => {
        eventSource.close();
      };
    });
  }
}
