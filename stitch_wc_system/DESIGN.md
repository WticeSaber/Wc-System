---
name: Obsidian Telemetry
colors:
  surface: '#0f131d'
  surface-dim: '#0f131d'
  surface-bright: '#353944'
  surface-container-lowest: '#0a0e18'
  surface-container-low: '#171b26'
  surface-container: '#1c1f2a'
  surface-container-high: '#262a35'
  surface-container-highest: '#313540'
  on-surface: '#dfe2f1'
  on-surface-variant: '#b9cbb9'
  inverse-surface: '#dfe2f1'
  inverse-on-surface: '#2c303b'
  outline: '#849585'
  outline-variant: '#3b4b3d'
  surface-tint: '#00e478'
  primary: '#f1ffef'
  on-primary: '#003919'
  primary-container: '#00ff87'
  on-primary-container: '#007138'
  inverse-primary: '#006d36'
  secondary: '#bdf4ff'
  on-secondary: '#00363d'
  secondary-container: '#00e3fd'
  on-secondary-container: '#00616d'
  tertiary: '#fffaf9'
  on-tertiary: '#67001d'
  tertiary-container: '#ffd4d6'
  on-tertiary-container: '#c01542'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#60ff98'
  primary-fixed-dim: '#00e478'
  on-primary-fixed: '#00210c'
  on-primary-fixed-variant: '#005227'
  secondary-fixed: '#9cf0ff'
  secondary-fixed-dim: '#00daf3'
  on-secondary-fixed: '#001f24'
  on-secondary-fixed-variant: '#004f58'
  tertiary-fixed: '#ffdadb'
  tertiary-fixed-dim: '#ffb2b8'
  on-tertiary-fixed: '#40000f'
  on-tertiary-fixed-variant: '#91002d'
  background: '#0f131d'
  on-background: '#dfe2f1'
  surface-variant: '#313540'
typography:
  display-xl:
    fontFamily: Space Grotesk
    fontSize: 20px
    fontWeight: '900'
    lineHeight: 28px
    letterSpacing: -0.02em
  title-lg:
    fontFamily: Space Grotesk
    fontSize: 18px
    fontWeight: '700'
    lineHeight: 24px
    letterSpacing: -0.01em
  body-base:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  body-bold:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '700'
    lineHeight: 20px
  stat-xl:
    fontFamily: JetBrains Mono
    fontSize: 30px
    fontWeight: '900'
    lineHeight: 36px
    letterSpacing: -0.03em
  stat-lg:
    fontFamily: JetBrains Mono
    fontSize: 24px
    fontWeight: '800'
    lineHeight: 30px
    letterSpacing: -0.02em
  label-mono:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
    letterSpacing: 0.05em
  label-mono-mobile:
    fontFamily: JetBrains Mono
    fontSize: 11px
    fontWeight: '500'
    lineHeight: 14px
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  unit: 4px
  gutter: 24px
  margin-page: 32px
  padding-card: 20px
  gap-compact: 12px
---

## Brand & Style

The visual identity is anchored in a high-stakes, technical aesthetic reminiscent of a tactical command room or an algorithmic trading floor. It is designed for analytical precision, projecting a mood of computational authority and modern efficiency.

The design system employs a **Corporate / Modern** framework with heavy influences from **Glassmorphism** and **High-Contrast** movements. This results in a "Cyberpunk Telemetry" style—utilizing deep obsidian surfaces, sharp hairlines, and high-energy neon glows to facilitate intense focus on dense data matrices. The hierarchy is strictly disciplined, ensuring that radiant primary indicators pop against a low-glare, dark environment to maintain legibility in professional, data-heavy contexts.

## Colors

The palette is structured into functional semantic tiers, separating deep structural backgrounds from vibrant interactive highlights.

### Foundations
- **Background**: Deep Space Obsidian (`#0B0F19`) provides a matte, low-glare base.
- **Surface**: Obsidian Card Base (`#161A23`) defines primary containers and dashboard cards.
- **Sunken**: Console Input Surface (`#111622`) is used for fields and tracks to create recessed depth.
- **Interactive Surface**: Slate Hover (`#28324A`) indicates active selection states.

### Accents
- **Primary (Home/Positive)**: Fluorescent Green (`#00FF87`) for primary CTAs and home-team victory paths.
- **Secondary (Info/Neutral)**: Technical Cyan (`#00E5FF`) for informational badges and draw pathways.
- **Tertiary (Away/Negative)**: Coral Red (`#FF4A6B`) for away-team victory paths and critical warnings.

### Typography
- **Primary**: High-Contrast White (`#FFFFFF`) for titles and active values.
- **Secondary**: Light Technical Gray (`#D1D5DB`) for labels and body text.
- **Muted**: Slate Gray (`#9CA3AF`) for secondary metadata and placeholders.

## Typography

The typography system handles extreme information density while maintaining mathematical precision.

- **Headlines (Space Grotesk)**: Geometric and industrial. Used for branding and major section titles to project a technical personality.
- **Interface (Inter)**: High legibility and neutral structure. Used for all labels, control descriptions, and primary body copy.
- **Data (JetBrains Mono)**: Monospaced to ensure numerical values, scores, and percentages align perfectly in tables and grids, preventing visual "jitter" during real-time updates.

All technical labels and code-strings should utilize `uppercase` with wide letter-spacing to enhance professional scannability.

## Layout & Spacing

This design system uses a **Fluid Grid** model optimized for high-resolution widescreen terminals. 

- **Grid Strategy**: A 12-column global grid. Standard dashboard layouts use an asymmetric split: a 3-column (25%) sidebar for "Control Console" inputs and a 9-column (75%) "Predictions Board" for data visualization.
- **Rhythm**: All spacing is derived from a 4px base unit. 
- **Breakpoints**: 
    - **Desktop (1024px+)**: Sidebar is fixed/sticky on the left with a 24px gutter.
    - **Mobile/Tablet (<1024px)**: Layout reflows into a single vertical column. The 6x6 heatmap matrix enables horizontal `overflow-auto` with a `min-width: 480px` to preserve data integrity on small screens.

## Elevation & Depth

Visual hierarchy is established through **Tonal Layers** and **Ambient Glows** rather than traditional drop shadows.

- **Layering**: The background is the lowest tier (`#0B0F19`). Cards (`#161A23`) sit on top with a thin hairline border (`#232A3B`) for definition.
- **Glows**: Critical elements (like the 1st place podium or primary buttons) use neon-tinted ambient shadows (`rgba(0, 255, 135, 0.35)`) to simulate light emission.
- **Sunken Elements**: Inputs and tracks use a darker fill (`#111622`) than their parent card to appear recessed into the interface.
- **Interactions**: Selected states trigger a border-color shift to neon green or cyan, paired with a soft backdrop blur to separate the focused element from the surrounding noise.

## Shapes

The shape language balances modern software aesthetics with technical rigor.

- **Standard Elements**: UI elements like inputs and heatmap cells use a standard `0.5rem` (8px) radius.
- **Containers**: Primary dashboard cards use a more generous `rounded-2xl` (16px) to soften the density of the information within.
- **Primary CTAs**: Large action buttons use `rounded-xl` (12px) to feel substantial and tactile.
- **Status Badges**: Small metadata chips use pill-shapes (full rounding) to contrast against the geometric grid.

## Components

### Buttons
- **Primary Action**: Solid `#00FF87` background with `#000000` text. Features a persistent radiant glow and a tactile `scale-95` on click.
- **Secondary Selectors**: Dark grey-blue (`#1F2638`) background with a `#2D3A56` border. Hovering triggers a neon-green or cyan border glow depending on the team context.

### Cards & Podium
- **Standard Card**: Obsidian base with a 1px border. 
- **Podium Cards**: A specialized set of 3 cards with varying heights. The center (1st place) card is the tallest, featuring a `2px` neon green border and a glowing back-halo effect.

### Input Fields & Sliders
- **Numerical Inputs**: Sunken dark background with monospaced text. Focus state removes the browser ring in favor of a solid neon green border.
- **Technical Sliders**: Low-profile tracks (`6px` height) with large circular thumbs color-coded to the team (Green for Home, Cyan for Away).

### Data Visualization
- **Heatmap Matrix**: 6x6 grid of cells. Background opacity of `#00FF87` scales linearly based on probability value (4% to 85%). The "Target" cell (highest value) is highlighted with a neon ring and a pulsing star indicator.
- **Arch Gauges**: Semi-circular SVG tracks using semantic accent colors for segmented distribution (Win/Draw/Loss).