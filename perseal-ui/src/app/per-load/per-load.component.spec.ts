import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { PerLoadComponent } from './per-load.component';

describe('PerLoadComponent', () => {
  let component: PerLoadComponent;
  let fixture: ComponentFixture<PerLoadComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ PerLoadComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PerLoadComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
