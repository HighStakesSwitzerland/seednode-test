import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-monitor',
  templateUrl: './monitor.component.html',
  styleUrls: ['./monitor.component.css']
})
export class MonitorComponent implements OnInit {

  public dataSource: any;
  public displayedColumns = ["nodeId", "moniker", "status", "result"];

  constructor() { }

  ngOnInit(): void {
  }

}
