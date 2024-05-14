pkgname=e1s-git
_pkgname=e1s
pkgver=v1.0.34
pkgrel=1
pkgdesc='A package for running e1s application'
arch=(x86_64)
url='https://github.com/keidarcy/e1s'
license=(MIT)
depends=(go)
provides=(e1s)
conflicts=(e1s)
source=("git+$url")
sha256sums=('SKIP')

pkgver() {
  cd "$srcdir/$_pkgname"
  ( set -o pipefail
    git describe --long 2>/dev/null | sed 's/\([^-]*-g\)/r\1/;s/-/./g' ||
    printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
  )
}

build() {
  cd "$srcdir/$_pkgname"
  go build -o $_pkgname ./cmd/e1s/main.go
}

package() {
  cd "$srcdir/$_pkgname"
  install -Dm755 $_pkgname "$pkgdir/usr/bin/$_pkgname"
  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$_pkgname/LICENSE"
}
