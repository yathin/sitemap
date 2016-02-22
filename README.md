# sitemap

sitemap is a small utility to crawl a website and display its sitemap as a tree. It is written in Go primarily as a learning exercise.

## Usage
```
usage: ./sitemap site restrict-to-domain max-depth
```
Where **site** must be valid absolute URL, **restrict-to-domain** must be a boolean value and **max-depth** must be an integer greater than or equal to 1.
### Example
```
$ ./sitemap http://www.yathin.com true 2
Output: Path (Number of Scripts, Number of Files (e.g, CSS), Number of Images, Number of External Links)
www.yathin.com. (0, 0, 1, 5)
    www.yathin.com/portfolio/. (7, 4, 18, 6)
    www.yathin.com/wordpress/. (0, 0, 0, 125)
    www.yathin.com/wordpress/about/. (0, 0, 1, 38)
```
