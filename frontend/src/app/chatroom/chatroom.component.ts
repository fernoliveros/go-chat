import { Component } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { MessageService } from '../services/message.service';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-chatroom',
  templateUrl: './chatroom.component.html',
  styleUrls: ['./chatroom.component.scss'],
  standalone: false,
})
export class ChatroomComponent {
  chatForm = new FormGroup({
    message: new FormControl('', Validators.required),
  });

  private sendSub: Subscription = new Subscription();

  public messages: { id: number; message: string }[] = [];

  constructor(private messageService: MessageService) {}

  ngOnInit() {
    console.log('Chat room initialized');
    this.messageService.getMessagesSSE().subscribe({
      next: (value) => {
        let counter = 1;
        this.messages = value.data.split(',').map((it: string) => {
          return {
            id: counter++,
            message: it,
          };
        });
      },
      error: (err) => {
        console.error('SSE receive message error', err);
      },
      complete: () => {
        console.log('SSE message observable completed');
      },
    });
  }

  sendMessage() {
    this.sendSub = this.messageService
      .sendMessage(this.chatForm.value)
      .subscribe({
        next: (value) => {
          console.log('successfully sent message', value);
          this.chatForm.reset();
        },
        error: (err) => {
          console.error('Send message error', err);
        },
        complete: () => {
          console.log('Send message observable completed');
        },
      });
  }

  ngOnDestroy(): void {
    console.log('Chat room destroyed!!!');
    if (this.sendSub) {
      this.sendSub.unsubscribe();
    }
  }
}
