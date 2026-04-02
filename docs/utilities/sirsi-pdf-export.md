Sirsi PDF Export Utility
Version 1.0.0

A standalone JavaScript utility for exporting HTML slide decks to PDF. Uses html2canvas for pixel-perfect capture and jsPDF for PDF generation. Dependencies are auto-loaded from CDN on first use.

Built and battle-tested on the Sirsi Technologies investor pitch deck (15 slides, complex layouts, CSS grids, gradients, Font Awesome icons, base64 images).


INSTALLATION

Add a single script tag to any HTML page:

    <script src="sirsi-pdf-export.js" defer></script>

The file lives at: packages/sirsi-ui/scripts/sirsi-pdf-export.js
Copy it to any project that needs HTML-to-PDF export.


USAGE

Minimal:

    SirsiPdfExport.exportSlides({
        filename: 'My_Document.pdf'
    });

Full configuration:

    SirsiPdfExport.exportSlides({
        containerSelector: '.presentation',
        slideSelector: '.slide',
        hideSelector: '.print-hide',
        pageWidth: 1440,
        pageHeight: 810,
        orientation: 'landscape',
        filename: 'Sirsi_Technologies_Pitch_Deck.pdf',
        imageQuality: 0.92,
        backgroundColor: '#ffffff',
        captureDelay: 500,
        showDefaultOverlay: true,
        overlayStyle: {
            fontFamily: "'Cinzel', serif",
            accentColor: '#C8A951',
            backgroundColor: 'rgba(0,0,0,0.85)'
        },
        showSlide: function(idx) { showSlide(idx + 1); },
        beforeCapture: function(idx, slideEl) { /* async prep */ },
        onProgress: function(current, total) { console.log(current + '/' + total); }
    });


API

SirsiPdfExport.exportSlides(options) -> Promise<void>

Options:

    containerSelector   CSS selector for the slide container          default: '.presentation'
    slideSelector       CSS selector for individual slides            default: '.slide'
    hideSelector        Elements to remove from clone before capture  default: '.print-hide'
    pageWidth           PDF page width in pixels                      default: 1440
    pageHeight          PDF page height in pixels                     default: 810
    orientation         'landscape' or 'portrait'                     default: 'landscape'
    filename            Output PDF filename                           default: 'export.pdf'
    imageQuality        JPEG quality 0-1                              default: 0.92
    backgroundColor     Canvas/page background color                  default: '#ffffff'
    captureDelay        ms to wait after showSlide before capture     default: 500
    showDefaultOverlay  Show built-in progress overlay                default: true
    overlayStyle        { fontFamily, accentColor, backgroundColor }  default: Inter/gold/dark
    showSlide           function(index) to navigate slides            default: null (clone-only)
    beforeCapture       async function(index, slideElement)           default: null
    onProgress          function(current, total)                      default: null


CAPTURE STRATEGY

The utility tries four strategies per slide, stopping at the first success:

    1. scale 1.5, allowTaint true    (best quality, may fail on cross-origin resources)
    2. scale 1.5, allowTaint false   (skips tainted resources but keeps quality)
    3. scale 1.0, allowTaint true    (smaller canvas, less memory pressure)
    4. scale 1.0, allowTaint false   (safest fallback)

If all four fail, falls back to a document.body capture. If that also fails, generates a placeholder slide with an error message.


KNOWN PITFALLS (html2canvas)

These were discovered during production testing and are documented here to save future teams the debugging time.

1. radial-gradient with background-size
   Elements using background-image: radial-gradient() with background-size create repeating patterns.
   html2canvas converts these to canvas patterns via createPattern(). If the element height or width
   rounds to 0px at the capture scale, createPattern throws "image argument is a canvas element
   with a width or height of 0". Fix: avoid gradient backgrounds on elements smaller than 2px.
   Use border-top or border-bottom instead of 1px gradient dividers.

2. Cross-origin images
   Any <img> loaded from a different origin (including file:// protocol for local files) taints the
   canvas, causing toDataURL to throw SecurityError. Fix: inline images as base64 data URIs, or
   serve from the same origin with proper CORS headers.

3. CSS custom properties in inline styles
   var(--token) references in style="" attributes may not resolve in html2canvas cloned DOM.
   The :root variables exist in the stylesheet but the clone may not inherit them correctly.
   Fix: use resolved hex/rgba values in inline styles. CSS classes with var() references work fine.

4. clip-path
   Partially supported. Complex polygon paths may cause the entire element content to disappear
   from the capture. Fix: add class="print-hide" and provide a non-clipped fallback, or remove
   clip-path for PDF-targeted elements.

5. backdrop-filter
   Not supported. Glass/blur effects render as solid backgrounds.

6. Web fonts from CDN
   Font Awesome and Google Fonts load asynchronously. If capture runs before fonts load, icons
   render as squares or missing glyphs. The 500ms captureDelay helps. For critical renders,
   use document.fonts.ready before calling exportSlides.

7. Canvas elements (Chart.js)
   html2canvas clones the DOM but not canvas pixel data. Chart.js canvases in the clone will be
   blank. Fix: use the beforeCapture hook to redraw charts, or convert chart canvases to <img>
   elements before export.

8. Large canvases
   Browsers silently fail when canvas dimensions exceed ~16384px in either direction. At scale 2,
   a 1440px-wide slide produces a 2880px canvas (safe). At scale 3, it would be 4320px (still safe).
   But very large viewports or high DPI could exceed limits. Scale 1.5 is the recommended maximum.


TROUBLESHOOTING

Blank slides:
    Check browser console for createPattern or SecurityError messages.
    Most likely cause: gradient on a tiny element or a cross-origin image.

Partial captures:
    Element overflow:hidden may clip content in the clone. Add overflow:visible in a
    beforeCapture hook for the affected slide.

Wrong slide captured:
    The onclone callback hides all slides and shows only the target. If your slides use
    JavaScript-driven layout (height calculations, chart rendering), pass a showSlide
    function so the live DOM is also set to the correct slide before capture.

Missing icons:
    Font Awesome icons require the webfont to be loaded. Increase captureDelay or check
    document.fonts.ready. Alternatively, replace icons with Unicode characters or inline SVG.


PORTFOLIO INTEGRATION

This utility is designed for use across all Sirsi portfolio applications:
    SirsiNexusApp    Pitch deck, admin portal, investor materials
    FinalWishes      Estate planning document exports
    Assiduous        Property report exports

Source of truth: packages/sirsi-ui/scripts/sirsi-pdf-export.js
