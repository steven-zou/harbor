import { Component, OnInit, Output, EventEmitter } from '@angular/core';
import { DistributionService } from '../distribution.service';
import { DistributionProvider } from '../distribution-provider';

@Component({
  selector: 'dist-providers',
  templateUrl: './distribution-provider-drivers.component.html',
  styleUrls: ['./distribution-provider-drivers.component.scss']
})
export class DistributionProviderDriversComponent implements OnInit {

  providers: DistributionProvider[] = [];
  @Output()
  setupEvt: EventEmitter<DistributionProvider> = new EventEmitter<DistributionProvider>();

  constructor(private distService: DistributionService) { }

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.distService.getProviderDrivers()
    .subscribe(
      providers => this.providers = providers,
      err => console.error(err)
    );
  }

  createInstance(provider: DistributionProvider) {
    this.setupEvt.emit(provider);
  }

}
