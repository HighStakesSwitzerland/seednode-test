import {AfterViewInit, Component, OnDestroy, OnInit, ViewChild} from "@angular/core";
import {MatDialog} from "@angular/material/dialog";
import {MatPaginator} from "@angular/material/paginator";
import {MatSort} from "@angular/material/sort";
import {MatTableDataSource} from "@angular/material/table";
import {catchError, EMPTY, empty, filter, Subject, switchMap, takeUntil, timer} from "rxjs";
import {get as _get, isEmpty as _isEmpty} from "lodash-es";
import {Peer} from "../../lib/domain/peer";
import {SeedService} from "../../lib/infra/seed.service";
import {DialogComponent, DialogData} from "../dialog/dialog.component";

@Component({
  selector: "app-monitor",
  templateUrl: "./monitor.component.html",
  styleUrls: ["./monitor.component.css"]
})
export class MonitorComponent implements OnInit, AfterViewInit, OnDestroy {
  @ViewChild(MatPaginator)
  paginator!: MatPaginator;
  @ViewChild(MatSort)
  sort!: MatSort;

  public dataSource: MatTableDataSource<Peer> = new MatTableDataSource();
  public displayedColumns = ["nodeId", "moniker", "status", "LastValue", "peers"];

  getPeers = timer(0, 5000);
  _destroy$ = new Subject<void>()

  constructor(private readonly _seedService: SeedService,
              private readonly _dialog: MatDialog) {
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
    this.dataSource.sortingDataAccessor = _get;
  }

  ngOnInit(): void {
    this.getPeers.pipe(
      switchMap(() =>
        this._seedService.getAllPeers()
          .pipe(
            catchError(err => {
              if (this._dialog.openDialogs?.length === 0) {
                this.openDialog(err);
              }
              return EMPTY;
            }),
          )
      ),
      filter(value => !_isEmpty(value)),
      takeUntil(this._destroy$)
    ).subscribe(list => {
      this.dataSource.data = list;
    });
  }

  openDialog(err: any) {
    const dialogRef = this._dialog.open(DialogComponent, {
      data: {
        title: "Error",
        errors: err,
        httpError: err
      } as DialogData
    });
    dialogRef.afterClosed().subscribe();
  }

  ngOnDestroy(): void {
    this._destroy$.complete();
  }


}
