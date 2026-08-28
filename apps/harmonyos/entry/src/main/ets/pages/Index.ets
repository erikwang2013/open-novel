import { HomeTab } from './HomeTab';
import { BooksTab } from './BooksTab';
import { MineTab } from './MineTab';

@Entry
@Component
struct Index {
  @State currentIndex: number = 0;

  build() {
    Tabs({ barPosition: BarPosition.End, index: this.currentIndex }) {
      TabContent() {
        HomeTab()
      }
      .tabBar($r('app.string.home'))

      TabContent() {
        BooksTab()
      }
      .tabBar($r('app.string.allBooks'))

      TabContent() {
        MineTab()
      }
      .tabBar($r('app.string.mine'))
    }
    .onChange((index: number) => {
      this.currentIndex = index;
    })
    .backgroundColor($r('app.color.start_window_background'))
  }
}
