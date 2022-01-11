import {Component, ElementRef, OnInit, ViewChild} from "@angular/core";
import {SeedService} from "../../lib/infra/seed.service";
import {isNil} from "lodash-es";

@Component({
  selector: "app-seed-list",
  templateUrl: "./seed-list.component.html",
  styleUrls: ["./seed-list.component.css"]
})
export class SeedListComponent {

  @ViewChild("textarea")
  public textarea!: ElementRef;

  constructor(private readonly _seedService: SeedService) {
  }

  submit() {
    const textValue = this.textarea.nativeElement.value;
    let seedList: string[];
    if (isNil(textValue)) {
        alert("empty");
    }
    seedList = textValue.trim().split(/[,\r\n]+/);

    if (seedList?.length == 0) {
      alert("problem");
      return;
    }

    this._seedService.postSeedList(seedList).subscribe({
      next: (res) => alert(res),
      complete: () => console.log("OK"),
      error: (err) => alert(err.message)
    });
  }
}
