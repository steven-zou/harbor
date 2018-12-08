import { Observable } from 'rxjs';
import { MessageHandlerService } from './../../shared/message-handler/message-handler.service';
import { Component, OnInit, Output, EventEmitter, OnDestroy } from '@angular/core';
import { ProviderInstance } from '../distribution-provider';
import { DistributionService } from '../distribution.service';
import { State } from 'clarity-angular';
import { Subscription } from 'rxjs';
import { MsgChannelService } from '../msg-channel.service';

@Component({
  selector: 'dist-instances',
  templateUrl: './distribution-provider-instances.component.html',
  styleUrls: ['./distribution-provider-instances.component.scss']
})
export class DistributionProviderInstancesComponent implements OnInit, OnDestroy {

  loading: boolean = false;
  instances: ProviderInstance[] = [];
  pageSize: number = 25;
  totalCount: number = 0;
  lastFilteredKeyword: string = "";
  periodicalSubscription: Subscription;
  chanSub: Subscription;

  @Output()
  createEvt: EventEmitter<string> = new EventEmitter<string>();
  @Output()
  editEvt: EventEmitter<ProviderInstance> = new EventEmitter<ProviderInstance>();

  constructor(
    private disService: DistributionService,
    private msgHandler: MessageHandlerService,
    private chanService: MsgChannelService,
    ) { }

  ngOnInit() {
    this.loadData(null);
    this.chanSub = this.chanService.subscribe(function(msg:string){
      if (msg=="created" || msg == "updated"){
        this.loadData(null);
      }

      console.error("unknown msg")
    });
    /*this.periodicalSubscription = Observable.interval(5000).subscribe(x => {
      this.loadData(null);
    });*/
  }

  ngOnDestroy() {
    if (this.periodicalSubscription){
      this.periodicalSubscription.unsubscribe();
    }

    if (this.chanSub){
      this.chanSub.unsubscribe();
    }
  }

  loadData(st: State) {
    this.disService.getProviderInstances()
    .subscribe(
      instances => {
        this.instances = instances;
        this.totalCount = this.instances.length;
      },
      err => this.msgHandler.error(err)
    );
  }

  refresh() {
    this.loadData(null);
  }

  doFilter($evt: any){
    console.log($evt);
  }

  createInstance() {
    this.createEvt.emit("create");
  }

  enableInstance(ID: string) {
      let instance = {
        enabled: true
      };
      this.disService
        .updateProviderInstance(ID, instance)
        .subscribe(
          () => {
          this.msgHandler.info(`Instance ${ID} enabled`);
          this.loadData(null);
        },
        err => this.msgHandler.error(err));
  }

  disableInstance(ID: string) {
    let instance = {
      enabled: false
    };
    this.disService
      .updateProviderInstance(ID, instance)
      .subscribe(
        () => {
        this.msgHandler.info(`Instance ${ID} disabled`);
        this.loadData(null);
      }, 
      err => this.msgHandler.error(err));
  }

  deleteInstance(ID: string) {
    this.disService.deleteProviderInstance(ID).subscribe(
      res => {
        let reply: string = JSON.stringify(res);
        this.msgHandler.info(`Instance ${ID} deleted: ${reply}`);
        this.loadData(null);
      },
      err => this.msgHandler.error(err)
    )
  }

  editInstance(inst: ProviderInstance) {
    this.editEvt.emit(inst);
  }

  fmtTime(time: number) {
    let date = new Date();
    return date.setTime(time * 1000);
  }
}
