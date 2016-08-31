# WOW.js Integration

Motion UI can be paired with [WOW.js](http://mynameismatthieu.com/WOW/) to trigger animations as elements scroll into view. **[Here's a CodePen that illustrates the concepts below](http://codepen.io/gakimball/pen/WrKRRy)**.

To start, load the JavaScript for WOW.js. The quickest way to do this is by loading from a CDN:

```html
<script src="https://cdnjs.cloudflare.com/ajax/libs/wow/1.1.2/wow.min.js"></script>
```

Next, in your main JavaScript file, initialize WOW:

```js
var wow = new WOW();
wow.init();
```

Lastly, we need animation classes to add to elements. Because Motion UI is a transition-focused library, there aren't many animation classes that come out of the box. The built-in animation classes are `.wiggle`, `.shake`, `.spin-cw`, and `.spin-ccw`. However, creating out own animation class using any of the transition effects is easy, using Motion UI's Sass mixins.

Here's a basic fade class. Refer to the [documentation on animations](animations.md) to learn more about how animations are built.

```scss
.animate-fade-in {
  @include mui-animation(fade);
}
```

Now we can apply this class to any element. We also add the class `.wow`, so WOW knows which elements to target as the page scrolls.

```html
<img class="wow animate-fade-in" src="//placekitten.com/300/300">
```



