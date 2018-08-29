import { Component, OnInit } from '@angular/core';
import { DistributionHistory } from '../distribution-history';
import { DistributionService } from '../distribution.service';
import { State } from 'clarity-angular';

@Component({
  selector: 'app-distribution-history',
  templateUrl: './distribution-history.component.html',
  styleUrls: ['./distribution-history.component.scss']
})
export class DistributionHistoryComponent implements OnInit {

  loading: boolean = false;
  records: DistributionHistory[] = [];
  pageSize: number = 50;
  totalCount: number = 0;
  lastFilteredKeyword: string = "";

  constructor(private disService: DistributionService) { }

  ngOnInit() {
  }

  refresh(st: State) {
    this.disService.getDistributionHistories()
    .subscribe(
      histories => {
        this.records = histories;
        this.totalCount = this.records.length;
      },
      err => console.error(err)
    );
  }

  doFilter($evt){
    console.log($evt);
  }

}
