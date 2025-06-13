import cairo

print("Starting cairo SVG creation...")

with cairo.SVGSurface("example.svg", 200, 200) as surface:
    print("✓ SVG surface created successfully")
    
    context = cairo.Context(surface)
    print("✓ Cairo context created")
    
    x, y, x1, y1 = 0.1, 0.5, 0.4, 0.9
    x2, y2, x3, y3 = 0.6, 0.1, 0.9, 0.5
    
    context.scale(200, 200)
    context.set_line_width(0.04)
    
    # Draw the curve
    context.move_to(x, y)
    context.curve_to(x1, y1, x2, y2, x3, y3)
    context.stroke()
    print("✓ Bezier curve drawn")
    
    # Draw the control lines
    context.set_source_rgba(1, 0.2, 0.2, 0.6)
    context.set_line_width(0.02)
    context.move_to(x, y)
    context.line_to(x1, y1)
    context.move_to(x2, y2)
    context.line_to(x3, y3)
    context.stroke()
    print("✓ Control lines drawn")

print("✓ SVG file 'example.svg' created successfully!")
print(f"  Cairo version: {cairo.version}")
print(f"  PyCairo version: {cairo.version_info}")
