import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { DistributionSetupModalComponent } from './distribution-setup-modal.component';

describe('DistributionSetupModalComponent', () => {
  let component: DistributionSetupModalComponent;
  let fixture: ComponentFixture<DistributionSetupModalComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DistributionSetupModalComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DistributionSetupModalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
