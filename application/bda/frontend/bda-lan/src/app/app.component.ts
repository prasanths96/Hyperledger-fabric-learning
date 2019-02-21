import { Component, OnInit, NgZone } from '@angular/core';
import { AuthService } from './services/auth.service';
import { Observable } from 'rxjs'
import { Http, Response } from '@angular/http';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit {
  title = 'bda-lan';
  private sse : any;
  private record:Observable<any>;

  constructor(private authService: AuthService,
    private zone:NgZone) {
      EventSource=window['EventSource'];
      this.sse=new EventSource('http://10.199.1.143:8000/api/data');
    }


  ngOnInit() {
    this.authService.autoAuthUser();
    this.record=this.getMessages();
    this.sse.onmessage=evt=> {
      console.log("EVTDATA:"+evt.data);
      // console.log(this.record);
      //this.zone.run(() =>observer.next(evt.data));
      //this.zone.run(() =>console.log(evt.data));
    };
  }
  signOut(){
    this.authService.logout();
  }

  getMessages():Observable<any> {
    return new Observable<any>(observer=> {
     
    this.sse.onmessage=evt=> {
      console.log("EVTDATA:"+evt.data);
      // console.log(this.record);
      //this.zone.run(() =>observer.next(evt.data));
      this.zone.run(() =>console.log(evt.data));
    };
    return () =>this.sse.close();
    });
  }

}
