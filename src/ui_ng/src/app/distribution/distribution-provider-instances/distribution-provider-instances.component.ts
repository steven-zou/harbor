import { Observable } from 'rxjs';
import { MessageHandlerService } from './../../shared/message-handler/message-handler.service';
import { Component, OnInit, Output, EventEmitter, OnDestroy } from '@angular/core';
import { ProviderInstance } from '../distribution-provider';
import { DistributionService } from '../distribution.service';
import { State } from 'clarity-angular';
import { Subscription } from 'rxjs';

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
  @Output()
  createEvt: EventEmitter<string> = new EventEmitter<string>();
  @Output()
  editEvt: EventEmitter<ProviderInstance> = new EventEmitter<ProviderInstance>();

  constructor(private disService: DistributionService,
    private msgHandler: MessageHandlerService) { }

  ngOnInit() {
    this.loadData(null);
    this.periodicalSubscription = Observable.interval(5000).subscribe(x => {
      this.loadData(null);
    });
  }

  ngOnDestroy() {
    this.periodicalSubscription.unsubscribe();
  }

  loadData(st: State) {
    this.disService.getProviderInstances()
    .subscribe(
      instances => {
        this.instances = instances;
        this.totalCount = this.instances.length;
      },
      err => console.error(err)
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
    console.log("enable ", ID);
      let instance = {
        enabled: true
      };
      this.disService
        .updateProviderInstance(ID, instance)
        .subscribe(res => {
          this.msgHandler.info('enable success');
          this.loadData(null);
        }
          , () => this.msgHandler.error);
  }

  disableInstance(ID: string) {
    console.log("disable ", ID);
    let instance = {
      enabled: false
    };
    this.disService
      .updateProviderInstance(ID, instance)
      .subscribe(res => {
        this.msgHandler.info('disable success');
        this.loadData(null);
      }, () => this.msgHandler.error);
  }

  deleteInstance(ID: string) {
    this.disService.deleteProviderInstance(ID).subscribe(
      () => this.msgHandler.info,
      () => this.msgHandler.handleError
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
