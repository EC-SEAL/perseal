import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { PerCodeComponent } from './per-code.component';

describe('PerCodeComponent', () => {
  let component: PerCodeComponent;
  let fixture: ComponentFixture<PerCodeComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ PerCodeComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PerCodeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
