import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DistributionHistoryComponent } from './distribution-history/distribution-history.component';
import { DistributionProvidersComponent } from './distribution-providers/distribution-providers.component';
import { DistributionSetupModalComponent } from './distribution-setup-modal/distribution-setup-modal.component';
import { DistributionService } from './distribution.service';
import { SharedModule } from "../shared/shared.module";

@NgModule({
  imports: [
    CommonModule,
    SharedModule
  ],
  declarations: [
    DistributionHistoryComponent, 
    DistributionProvidersComponent, 
    DistributionSetupModalComponent],
  providers: [DistributionService]
})
export class DistributionModule { }
