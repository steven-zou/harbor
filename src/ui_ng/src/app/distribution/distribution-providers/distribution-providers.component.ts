import { Component, OnInit, ViewChild } from '@angular/core';
import { DistributionSetupModalComponent } from '../distribution-setup-modal/distribution-setup-modal.component';
import { DistributionProvider, ProviderInstance } from '../distribution-provider';

@Component({
  selector: 'dist-providers-base',
  templateUrl: './distribution-providers.component.html',
  styleUrls: ['./distribution-providers.component.scss']
})
export class DistributionProvidersComponent implements OnInit {

  instanceActive: boolean = true;
  providerActive: boolean = false;
  @ViewChild("setupModal")
  setupModal: DistributionSetupModalComponent;

  constructor() { }

  ngOnInit() {
  }

  create(evt: string) {
    this.instanceActive = false;
    this.providerActive = true;
  }

  createInstance(provider: DistributionProvider) {
    this.setupModal.openSetupModal("create", provider);
  }

  editInstance(instance: ProviderInstance) {
    this.setupModal.openSetupModal("edit", instance);
  }
}
