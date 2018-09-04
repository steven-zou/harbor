import { Component, OnInit, Output, EventEmitter } from '@angular/core';
import { ProviderInstance } from '../distribution-provider';
import { DistributionService } from '../distribution.service';
import { State } from 'clarity-angular';

@Component({
  selector: 'dist-instances',
  templateUrl: './distribution-provider-instances.component.html',
  styleUrls: ['./distribution-provider-instances.component.scss']
})
export class DistributionProviderInstancesComponent implements OnInit {

  loading: boolean = false;
  instances: ProviderInstance[] = [];
  pageSize: number = 25;
  totalCount: number = 0;
  lastFilteredKeyword: string = "";
  @Output()
  createEvt: EventEmitter<string> = new EventEmitter<string>();
  @Output()
  editEvt: EventEmitter<ProviderInstance> = new EventEmitter<ProviderInstance>();

  constructor(private disService: DistributionService) { }

  ngOnInit() {
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

  refresh(){
    this.loadData(null);
  }

  doFilter($evt: any){
    console.log($evt);
  }

  createInstance(){
    this.createEvt.emit("create");
  }

  enableInstance(ID: string){
    console.log("enable ", ID)
  }

  disableInstance(ID: string){
    console.log("disable ", ID)
  }

  deleteInstance(ID: string){
    console.log("delete ", ID)
  }

  editInstance(inst: ProviderInstance){
    this.editEvt.emit(inst);
  }

}
