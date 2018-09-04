import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { DistributionProviderDriversComponent } from './distribution-provider-drivers.component';

describe('DistributionProviderDriversComponent', () => {
  let component: DistributionProviderDriversComponent;
  let fixture: ComponentFixture<DistributionProviderDriversComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DistributionProviderDriversComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DistributionProviderDriversComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
