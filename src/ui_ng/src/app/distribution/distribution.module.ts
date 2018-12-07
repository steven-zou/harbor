import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DistributionHistoryComponent } from './distribution-history/distribution-history.component';
import { DistributionProvidersComponent } from './distribution-providers/distribution-providers.component';
import { DistributionSetupModalComponent } from './distribution-setup-modal/distribution-setup-modal.component';
import { DistributionService } from './distribution.service';
import { SharedModule } from "../shared/shared.module";
import { DistributionProviderDriversComponent } from './distribution-provider-drivers/distribution-provider-drivers.component';
import { DistributionProviderInstancesComponent } from './distribution-provider-instances/distribution-provider-instances.component';
import { MsgChannelService } from './msg-channel.service';

@NgModule({
  imports: [
    CommonModule,
    SharedModule
  ],
  declarations: [
    DistributionHistoryComponent, 
    DistributionProvidersComponent, 
    DistributionSetupModalComponent, 
    DistributionProviderDriversComponent,
    DistributionProviderInstancesComponent],
  providers: [
    DistributionService,
    MsgChannelService,
  ]
})
export class DistributionModule { }
