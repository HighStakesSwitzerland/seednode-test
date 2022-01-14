import {HttpErrorResponse} from "@angular/common/http";
import {Component, ElementRef, EventEmitter, OnInit, Output, ViewChild} from "@angular/core";
import {MatDialog} from "@angular/material/dialog";
import {SeedService} from "../../lib/infra/seed.service";
import {isNil} from "lodash-es";
import {DialogComponent, DialogData} from "../dialog/dialog.component";

@Component({
  selector: "app-seed-list",
  templateUrl: "./seed-list.component.html",
  styleUrls: ["./seed-list.component.css"]
})
export class SeedListComponent {

  @ViewChild("textarea")
  public textarea!: ElementRef;

  @Output()
  public backToMonitorTab = new EventEmitter();

  public working: boolean = false;

  constructor(private readonly _seedService: SeedService,
              private readonly _dialog: MatDialog) {
  }

  submit() {
    const textValue = this.textarea.nativeElement.value;
    let seedList: string[];
    if (isNil(textValue)) {
        alert("empty");
    }
    seedList = textValue.trim().split(/[,\r\n]+/);

    if (seedList?.length == 0) {
      return;
    }

    this.working = true;
    this._seedService.postSeedList(seedList).subscribe({
      complete: () => {
        this.working = false;
        this.textarea.nativeElement.value = "";
        this.backToMonitorTab.next(undefined);
      },
      error: (err: HttpErrorResponse) => {
        this.working = false;
        this.textarea.nativeElement.value = "";
        this.openDialog(err);
      }
    });
  }

  openDialog(err: HttpErrorResponse) {
    const dialogRef = this._dialog.open(DialogComponent, {
      data: {
        title: "Error",
        errors: err.error,
        httpError: err
      } as DialogData
    });
    dialogRef.afterClosed().subscribe();
  }
}
