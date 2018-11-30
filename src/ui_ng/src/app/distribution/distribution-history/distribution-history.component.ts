import { Component, OnInit, OnDestroy } from '@angular/core';
import { DistributionHistory } from '../distribution-history';
import { DistributionService } from '../distribution.service';
import { State } from 'clarity-angular';
import { Observable, Subscription } from 'rxjs';

@Component({
  selector: 'app-distribution-history',
  templateUrl: './distribution-history.component.html',
  styleUrls: ['./distribution-history.component.scss']
})
export class DistributionHistoryComponent implements OnInit, OnDestroy {

  loading: boolean = false;
  records: DistributionHistory[] = [];
  pageSize: number = 50;
  totalCount: number = 0;
  lastFilteredKeyword: string = "";
  periodicalSubscription: Subscription;

  constructor(private disService: DistributionService) { }

  ngOnInit() {
    this.loadData(null);
    this.periodicalSubscription = Observable.interval(5000).subscribe(x => {
      this.loadData(null);
    });
  }

  ngOnDestroy(): void {
    //Called once, before the instance is destroyed.
    //Add 'implements OnDestroy' to the class.
    this.periodicalSubscription.unsubscribe();
  }

  loadData(st: State) {
    this.disService.getDistributionHistories()
    .subscribe(
      histories => {
        this.records = histories;
        this.totalCount = this.records.length;
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

}
