import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { DistributionProviderInstancesComponent } from './distribution-provider-instances.component';

describe('DistributionProviderInstanceComponent', () => {
  let component: DistributionProviderInstancesComponent;
  let fixture: ComponentFixture<DistributionProviderInstancesComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DistributionProviderInstancesComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DistributionProviderInstancesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
