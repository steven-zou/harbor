import { Component, OnInit } from '@angular/core';
import { ProviderInstance } from '../distribution-provider';
import { DistributionService } from '../distribution.service';
import { State } from 'clarity-angular';

@Component({
  selector: 'app-distribution-providers',
  templateUrl: './distribution-providers.component.html',
  styleUrls: ['./distribution-providers.component.scss']
})
export class DistributionProvidersComponent implements OnInit {

  loading: boolean = false;
  instances: ProviderInstance[] = [];
  pageSize: number = 25;
  totalCount: number = 0;
  lastFilteredKeyword: string = "";

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

  enableInstance(ID: string){
    console.log("enable ", ID)
  }

  disableInstance(ID: string){
    console.log("disable ", ID)
  }

  deleteInstance(ID: string){
    console.log("delete ", ID)
  }

  editInstance(ID: string){
    console.log("edit ", ID)
  }
}
