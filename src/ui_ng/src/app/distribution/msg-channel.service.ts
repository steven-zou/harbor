import { Injectable } from "@angular/core";
import { Subject } from 'rxjs/Subject';
import { Subscription } from 'rxjs/Subscription';

@Injectable()
export class MsgChannelService{
    chanSource = new Subject<string>();
    chanObs$ = this.chanSource.asObservable()

    public publish(msg:string){
        this.chanSource.next(msg);
    }

    public subscribe(callback: Function):Subscription{
        return this.chanObs$.subscribe((msg:string) =>{
            if (callback) {
                callback(msg);
            }
        });
    }
}