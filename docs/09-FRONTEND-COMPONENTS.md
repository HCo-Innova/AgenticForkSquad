# 09-FRONTEND-COMPONENTS.md

```markdown
# 09-FRONTEND-COMPONENTS.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** Frontend Components Specification  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), 
[08-API-SPECIFICATION.md](08-API-SPECIFICATION.md), 
[01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md)

---

## 📖 Table of Contents

1. [Frontend Architecture](#frontend-architecture)
2. [Component Organization](#component-organization)
3. [Layout Components](#layout-components)
4. [Task Components](#task-components)
5. [Agent Components](#agent-components)
6. [Optimization Components](#optimization-components)
7. [Consensus Components](#consensus-components)
8. [Common Components](#common-components)
9. [Custom Hooks](#custom-hooks)
10. [State Management](#state-management)
11. [Routing](#routing)
12. [Real-Time Updates](#real-time-updates)

---

## 🏗️ Frontend Architecture

### Technology Stack

**Core Framework:**
- React 18 (with Concurrent Features)
- TypeScript 5 (strict mode)
- Vite 5 (build tool, hot reload)

**Styling:**
- Tailwind CSS 3 (utility-first)
- PostCSS (processing)
- CSS Modules (component-scoped styles, optional)

**State Management:**
- React Query (TanStack Query v5) - Server state
- Context API - UI state (modals, theme, etc.)
- Local state (useState) - Component state

**HTTP Client:**
- Fetch API (native)
- React Query for caching and synchronization

**WebSocket:**
- Native WebSocket API
- Custom hook wrapper

**Code Quality:**
- ESLint (linting)
- Prettier (formatting)
- TypeScript (type safety)

---

### Architecture Principles

**Component-Based:**
- Small, focused components
- Single Responsibility Principle
- Reusable across features
- Max 300 lines per component file

**Container/Presentational Pattern:**
- Container: Logic, state, data fetching
- Presentational: Pure UI, receives props
- Clear separation of concerns

**Atomic Design (Modified):**
- Atoms: Buttons, Inputs, Icons
- Molecules: Cards, Forms, Lists
- Organisms: Sections, Layouts
- Pages: Route components

**Data Flow:**
- Top-down (props)
- Events bubble up (callbacks)
- Global state via Context
- Server state via React Query

---

### Directory Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── layout/          # App structure
│   │   ├── task/            # Task feature
│   │   ├── agent/           # Agent feature
│   │   ├── optimization/    # Optimization feature
│   │   ├── consensus/       # Consensus feature
│   │   └── common/          # Shared components
│   │
│   ├── hooks/               # Custom hooks
│   │   ├── useWebSocket.ts
│   │   ├── useTasks.ts
│   │   ├── useOptimizations.ts
│   │   └── useAgents.ts
│   │
│   ├── services/            # API clients
│   │   ├── api.ts
│   │   └── websocket.ts
│   │
│   ├── types/               # TypeScript types
│   │   ├── task.types.ts
│   │   ├── agent.types.ts
│   │   ├── optimization.types.ts
│   │   └── consensus.types.ts
│   │
│   ├── context/             # Context providers
│   │   ├── ThemeContext.tsx
│   │   └── NotificationContext.tsx
│   │
│   ├── utils/               # Utilities
│   │   ├── formatters.ts
│   │   ├── validators.ts
│   │   └── constants.ts
│   │
│   ├── pages/               # Route pages
│   │   ├── HomePage.tsx
│   │   ├── TaskListPage.tsx
│   │   ├── TaskDetailPage.tsx
│   │   └── AgentsPage.tsx
│   │
│   ├── App.tsx              # Root component
│   ├── main.tsx             # Entry point
│   └── index.css            # Global styles
│
├── public/                  # Static assets
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

---

## 🧩 Component Organization

### Feature-Based Structure

**By Feature Domain:**

**Task Feature:**
- TaskSubmission (create new task)
- TaskList (list all tasks)
- TaskDetail (single task view)
- TaskCard (compact task display)
- TaskStatusBadge (status indicator)

**Agent Feature:**
- AgentGrid (all agents overview)
- AgentCard (single agent display)
- AgentStatus (real-time status)
- AgentMetrics (performance stats)

**Optimization Feature:**
- ProposalList (all proposals for task)
- ProposalCard (single proposal)
- ProposalComparison (side-by-side view)
- BenchmarkChart (performance visualization)
- SQLViewer (syntax-highlighted SQL)

**Consensus Feature:**
- ConsensusVisualization (score comparison)
- ScoreBreakdown (detailed scoring)
- DecisionRationale (explanation text)
- WinnerAnnouncement (highlight winner)

---

### Naming Conventions

**Component Files:**
- PascalCase: `TaskSubmission.tsx`
- Descriptive: `ProposalComparison.tsx` not `Comparison.tsx`
- Feature prefix if ambiguous: `TaskCard.tsx` vs `AgentCard.tsx`

**Component Functions:**
- Match filename: `export function TaskSubmission()`
- Props interface: `TaskSubmissionProps`
- Internal functions: `handleSubmit`, `validateForm`

**Hooks:**
- Prefix `use`: `useTasks.ts`
- Descriptive: `useWebSocket.ts` not `useWS.ts`

**Types:**
- Suffix `.types.ts`: `task.types.ts`
- Interface names: `Task`, `TaskStatus`, `CreateTaskRequest`

---

## 🎨 Layout Components

### Layout.tsx

**Purpose:**  
Main application layout wrapper with header, sidebar, and content area.

**Responsibilities:**
- Render header with logo and navigation
- Render sidebar with menu items
- Render main content area
- Handle responsive layout (mobile/desktop)
- Provide consistent spacing and structure

**Structure:**
```
┌─────────────────────────────────────┐
│  Header                             │
│  - Logo                             │
│  - Navigation                       │
└─────────────────────────────────────┘
┌──────┬──────────────────────────────┐
│      │                              │
│ Side │  Main Content Area           │
│ bar  │  (children)                  │
│      │                              │
└──────┴──────────────────────────────┘
```

**Props:**
- `children`: React nodes (page content)

**State:**
- `sidebarOpen`: Boolean (mobile menu toggle)

**Responsive Behavior:**
- Desktop: Sidebar always visible
- Mobile: Sidebar collapses to hamburger menu
- Breakpoint: 768px (Tailwind `md:`)

---

### Header.tsx

**Purpose:**  
Top navigation bar with branding and main navigation links.

**Content:**
- Logo/Brand (left)
- Navigation links (center)
- System status indicator (right)
- User menu (right, future)

**Navigation Links:**
- Home (/)
- Tasks (/tasks)
- Agents (/agents)
- Stats (/stats, optional)

**System Status Indicator:**
- Green dot: All systems operational
- Yellow dot: Degraded performance
- Red dot: Service disruption
- Tooltip on hover with details

---

### Sidebar.tsx

**Purpose:**  
Side navigation menu with feature-specific links.

**Menu Structure:**
```
Dashboard
Tasks
  - New Task
  - All Tasks
  - Completed
  - Failed
Agents
  - Overview
  - Performance
Optimization History
Settings
```

**Active State:**
- Highlight current page
- Use React Router location
- Visual indicator (colored border/background)

**Collapsible Sections:**
- Click to expand/collapse
- Persist state in localStorage
- Smooth animation

---

## 📝 Task Components

### TaskSubmission.tsx

**Purpose:**  
Form for creating new optimization tasks.

**Form Fields:**

**Required:**
- Task Type (dropdown)
  - Options: Query Optimization, Schema Improvement, Index Recommendation
- Target Query (textarea)
  - Syntax highlighting (SQL)
  - Line numbers
  - Min 10 characters validation

**Optional:**
- Description (textarea)
  - Max 500 characters
  - Character counter
- Priority (radio buttons)
  - Options: Low, Medium, High
  - Default: Medium
- Advanced Options (accordion)
  - Max Storage (number input, MB)
  - Risk Tolerance (dropdown)
  - Custom Scoring Weights (sliders)

**State:**
- `formData`: Form values
- `errors`: Validation errors
- `isSubmitting`: Boolean (submit in progress)

**Validation:**
- Client-side: Required fields, format checks
- Real-time: Show errors on blur
- Submit-time: Comprehensive validation
- Display errors inline below fields

**Submission Flow:**
```
1. User clicks "Optimize Query"
2. Validate all fields
3. If invalid: Show errors, focus first error
4. If valid:
   - Disable form
   - Show loading spinner
   - POST to /api/v1/tasks
   - On success: Redirect to task detail page
   - On error: Show error message, re-enable form
```

**UX Enhancements:**
- Auto-save to localStorage (draft)
- Keyboard shortcuts (Cmd+Enter to submit)
- Query template suggestions
- Paste detection (auto-fill from clipboard)

---

### TaskList.tsx

**Purpose:**  
Display all tasks with filtering, sorting, and pagination.

**Layout:**
```
┌─────────────────────────────────────┐
│  Filters & Search                   │
│  [Status ▼] [Type ▼] [Search...]   │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│  TaskCard                           │
├─────────────────────────────────────┤
│  TaskCard                           │
├─────────────────────────────────────┤
│  TaskCard                           │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│  Pagination                         │
│  ← Previous  1 2 3 4 5  Next →     │
└─────────────────────────────────────┘
```

**Filters:**
- Status: All, Pending, In Progress, Completed, Failed
- Type: All, Query Optimization, Schema Improvement, Index Recommendation
- Search: Query text search (debounced)

**Sorting:**
- Created Date (newest/oldest)
- Completion Time
- Status

**State:**
- `filters`: Current filter values
- `sort`: Current sort field and direction
- `page`: Current page number
- `tasks`: Task data (from React Query)

**Data Fetching:**
- Use `useTasks` hook
- React Query handles caching
- Auto-refresh every 30 seconds
- Manual refresh button

**Empty States:**
- No tasks: "No tasks yet. Create your first optimization!"
- No results: "No tasks match your filters."

---

### TaskCard.tsx

**Purpose:**  
Compact task display for list views.

**Content:**
```
┌─────────────────────────────────────┐
│ [Status Badge]        [Type Badge]  │
│ Task #123                           │
│ "Optimize monthly revenue report"   │
│                                     │
│ SELECT u.email, SUM(o.total)...    │
│ (truncated query preview)           │
│                                     │
│ Created: 2 hours ago                │
│ Duration: 3m 45s                    │
└─────────────────────────────────────┘
```

**Props:**
- `task`: Task object
- `onClick`: Click handler (navigate to detail)

**Status Badge Colors:**
- Pending: Gray
- In Progress: Blue (animated pulse)
- Completed: Green
- Failed: Red

**Hover State:**
- Slight elevation (shadow)
- Cursor pointer
- Background color change

---

### TaskDetail.tsx

**Purpose:**  
Comprehensive view of single task with all details and sub-sections.

**Layout Sections:**

**1. Header:**
- Task ID and title
- Status badge (large)
- Created/Completed timestamps
- Action buttons (Retry if failed, Cancel if pending)

**2. Query Section:**
- Original query (syntax highlighted)
- Copy to clipboard button
- Expand/collapse (if long)

**3. Progress Timeline:**
- Visual timeline of workflow steps
- Current step highlighted
- Completed steps with checkmarks
- Future steps grayed out

**4. Agents Section:**
- Grid of agent cards
- Real-time status per agent
- Fork IDs displayed
- Duration per agent

**5. Proposals Section (Tab or Accordion):**
- List of all proposals
- Side-by-side comparison mode
- Expand to see full details

**6. Consensus Section:**
- Winner announcement (if decided)
- Score visualization (bar chart)
- Decision rationale (formatted text)
- Apply status (if optimization applied)

**State:**
- `task`: Task data (from React Query)
- `activeTab`: Current tab (Proposals, Consensus, Logs)
- `selectedProposals`: For comparison mode

**Real-Time Updates:**
- WebSocket connection for task events
- Update progress timeline
- Update agent statuses
- Show new proposals as they arrive

---

### TaskStatusBadge.tsx

**Purpose:**  
Visual indicator of task status.

**Props:**
- `status`: TaskStatus enum
- `size`: "small" | "medium" | "large"
- `animated`: Boolean (pulse for in-progress)

**Styles by Status:**
```
Pending:
  - Color: Gray (#6B7280)
  - Icon: Clock
  - No animation

In Progress:
  - Color: Blue (#3B82F6)
  - Icon: Spinner
  - Animated: Pulse

Completed:
  - Color: Green (#10B981)
  - Icon: Checkmark
  - No animation

Failed:
  - Color: Red (#EF4444)
  - Icon: X
  - No animation
```

**Accessibility:**
- ARIA label with status text
- Sufficient color contrast
- Icon + text (not color alone)

---

## 🤖 Agent Components

### AgentGrid.tsx

**Purpose:**  
Overview of all available agents with current status.

**Layout:**
```
┌──────────────┬──────────────┬──────────────┐
│ AgentCard    │ AgentCard    │ AgentCard    │
│ (gemini-2.5-pro) │ (gemini-2.5-flash) │ (gemini-2.0-flash) │
└──────────────┴──────────────┴──────────────┘
```

**Grid Properties:**
- 3 columns on desktop
- 2 columns on tablet
- 1 column on mobile
- Equal height cards
- Gap between cards

**Data Fetching:**
- `useAgents` hook
- Poll every 10 seconds (agents status changes)
- React Query caching

---

### AgentCard.tsx

**Purpose:**  
Display individual agent with status and capabilities.

**Content:**
```
┌─────────────────────────────────────┐
│ 🤖 gemini-2.5-pro (Vertex AI) │
│ [Status: Available]                 │
│                                     │
│ Specialization:                     │
│ • SQL optimization                  │
│ • Index design                      │
│ • Query rewriting                   │
│                                     │
│ Current Load: 1/3 tasks             │
│ Success Rate: 96%                   │
│ Avg Duration: 3m 15s                │
└─────────────────────────────────────┘
```

**Props:**
- `agent`: Agent object with metrics

**Status Indicator:**
- Available: Green dot
- Busy: Yellow dot (with task count)
- Unavailable: Red dot

**Metrics Display:**
- Progress bar for current load (1/3 = 33%)
- Success rate percentage
- Average duration formatted (3m 15s)

**Click Action:**
- Navigate to agent detail page
- Show recent tasks for this agent
- Performance charts

---

### AgentStatus.tsx

**Purpose:**  
Real-time status indicator for agent during task execution.

**Use Case:**  
Shown in TaskDetail to track agent progress.

**States:**

**Idle:**
```
┌─────────────────────────────────────┐
│ 🤖 gemini-2.5-pro    │
│ Status: Waiting to start            │
└─────────────────────────────────────┘
```

**Creating Fork:**
```
┌─────────────────────────────────────┐
│ 🤖 gemini-2.5-pro    │
│ Status: Creating fork...            │
│ [Progress spinner]                  │
└─────────────────────────────────────┘
```

**Analyzing:**
```
┌─────────────────────────────────────┐
│ 🤖 gemini-2.5-pro    │
│ Status: Analyzing query             │
│ Duration: 28s                       │
│ [Progress spinner]                  │
└─────────────────────────────────────┘
```

**Completed:**
```
┌─────────────────────────────────────┐
│ 🤖 gemini-2.5-pro    │
│ Status: ✓ Completed                │
│ Duration: 3m 15s                    │
│ Proposal: Partial Index             │
└─────────────────────────────────────┘
```

**Failed:**
```
┌─────────────────────────────────────┐
│ 🤖 gemini-2.5-pro    │
│ Status: ✗ Failed                   │
│ Error: LLM API timeout              │
└─────────────────────────────────────┘
```

**Props:**
- `agentExecution`: AgentExecution object
- `realTime`: Boolean (connect to WebSocket)

**Updates:**
- WebSocket events update status
- Duration updates every second
- Smooth transitions between states

---

### AgentMetrics.tsx

**Purpose:**  
Display historical performance metrics for an agent.

**Metrics Shown:**

**Success Rate:**
- Percentage (96%)
- Trend arrow (up/down)
- Sparkline chart (last 10 tasks)

**Average Duration:**
- Time formatted (3m 15s)
- Comparison to overall average
- Distribution histogram

**Win Rate (Consensus):**
- Percentage of proposals that won
- By task type breakdown
- Pie chart visualization

**Tasks Completed:**
- Total count
- This week/month/all time
- Line chart over time

**Props:**
- `agentType`: Agent identifier
- `timeRange`: "week" | "month" | "all"

**Charts:**
- Use Recharts or Chart.js
- Responsive sizing
- Tooltips on hover
- Color-coded by agent

---

## 🔧 Optimization Components

### ProposalList.tsx

**Purpose:**  
Display all proposals for a task.

**Layout:**
```
┌─────────────────────────────────────┐
│ 🥇 Winner: gemini-2.5-pro Proposal │
│ [ProposalCard - Expanded]           │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│ 🥈 Runner-up: gemini-2.5-flash Proposal │
│ [ProposalCard - Collapsed]          │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│ 🥉 Third: gemini-2.0-flash Proposal │
│ [ProposalCard - Collapsed]          │
└─────────────────────────────────────┘

[Compare Selected] button
```

**Features:**
- Winner highlighted (gold border, expanded by default)
- Click to expand/collapse
- Multi-select for comparison
- Sort by rank, agent, improvement

**Props:**
- `proposals`: Array of proposals
- `winnerId`: ID of winning proposal (optional)

**State:**
- `expandedIds`: Set of expanded proposal IDs
- `selectedIds`: Set for comparison

---

### ProposalCard.tsx

**Purpose:**  
Display single optimization proposal with details.

**Collapsed View:**
```
┌─────────────────────────────────────┐
│ gemini-2.5-pro • Partial Index │
│ 82.6% improvement • 12 MB overhead  │
│ [Expand ▼]                          │
└─────────────────────────────────────┘
```

**Expanded View:**
```
┌─────────────────────────────────────┐
│ gemini-2.5-pro • Partial Index │
│ 82.6% improvement • 12 MB overhead  │
│                                     │
│ SQL Commands:                       │
│ [SQLViewer component]               │
│                                     │
│ Rationale:                          │
│ "Partial index targets only..."    │
│                                     │
│ Benchmark Results:                  │
│ [BenchmarkChart component]          │
│                                     │
│ [Collapse ▲]                        │
└─────────────────────────────────────┘
```

**Props:**
- `proposal`: Proposal object
- `isWinner`: Boolean (highlight as winner)
- `expanded`: Boolean (controlled expansion)
- `onToggle`: Callback for expand/collapse

**Highlight Winner:**
- Gold/yellow border
- Trophy icon
- "Winner" badge
- Slight elevation

---

### ProposalComparison.tsx

**Purpose:**  
Side-by-side comparison of multiple proposals.

**Layout:**
```
┌──────────────┬──────────────┬──────────────┐
│ gemini-2.5-pro │ gemini-2.5-flash │ gemini-2.0-flash │
├──────────────┼──────────────┼──────────────┤
│ Partial Idx  │ Mat. View    │ Partitioning │
│              │              │              │
│ 82.6%        │ 93.5%        │ 61.5%        │
│ improvement  │ improvement  │ improvement  │
│              │              │              │
│ 12 MB        │ 80 MB        │ 40 MB        │
│ overhead     │ overhead     │ overhead     │
│              │              │              │
│ Score: 93.0  │ Score: 78.5  │ Score: 66.5  │
└──────────────┴──────────────┴──────────────┘
```

**Comparison Metrics:**
- Proposal type
- Improvement percentage (visual bar)
- Storage overhead (visual bar)
- Complexity (Low/Medium/High with color)
- Risk (Low/Medium/High with color)
- Final score (large number)

**Visual Encoding:**
- Green bars: Better values
- Red bars: Worse values
- Neutral: Gray
- Winner column highlighted

**Props:**
- `proposals`: Array of proposals to compare
- `maxItems`: Maximum columns (default 3)

---

### BenchmarkChart.tsx

**Purpose:**  
Visualize benchmark results (before/after comparison).

**Chart Type:**  
Grouped bar chart

**Data:**
```
Test 1 (Original):
  Baseline: 2,300 ms
  Optimized: 450 ms

Test 2 (Limited):
  Baseline: 800 ms
  Optimized: 120 ms

Test 3 (Filtered):
  Baseline: 1,200 ms
  Optimized: 180 ms

Test 4 (Sorted):
  Baseline: 2,400 ms
  Optimized: 460 ms
```

**Visual:**
```
        ┌─────────────────────────────────┐
2,400ms │   ███████████                   │ Baseline
        │   ████                          │ Optimized
        │                                 │
1,200ms │         ████████                │
        │         ██                      │
        │                                 │
    0ms └─────────────────────────────────┘
         Test1  Test2  Test3  Test4
```

**Props:**
- `benchmarkResults`: Array of results
- `showImprovement`: Boolean (show % labels)

**Features:**
- Hover tooltips with exact values
- Improvement percentage labels
- Color coding (blue for baseline, green for optimized)
- Responsive sizing

---

### SQLViewer.tsx

**Purpose:**  
Display SQL code with syntax highlighting.

**Features:**
- Syntax highlighting (keywords, strings, comments)
- Line numbers
- Copy to clipboard button
- Expand/collapse (if long)
- Read-only (not editable)

**Syntax Highlighting:**
- Use Prism.js or Monaco Editor
- SQL language support
- Theme: VS Code Dark (or light based on app theme)

**Props:**
- `sql`: SQL string (single command or array)
- `maxHeight`: Scroll if exceeds
- `showLineNumbers`: Boolean

**Example Display:**
```
1  CREATE INDEX idx_orders_user_completed
2  ON orders(user_id, status)
3  WHERE status = 'completed';
4
5  ANALYZE orders;

[Copy to Clipboard]
```

---

## ⚖️ Consensus Components

### ConsensusVisualization.tsx

**Purpose:**  
Visual representation of consensus decision with scores.

**Layout:**
```
┌─────────────────────────────────────┐
│           Consensus Decision         │
│                                     │
│  🥇 gemini-2.5-pro: 93.0 pts │
│  ████████████████████░░░ (93%)     │
│                                     │
│  🥈 gemini-2.5-pro: 78.5 pts        │
│  ███████████████░░░░░░░░ (79%)     │
│                                     │
│  🥉 gemini-2.0-flash: 66.5 pts │
│  █████████████░░░░░░░░░░ (67%)     │
└─────────────────────────────────────┘
```

**Visual Elements:**
- Horizontal bars showing relative scores
- Medal icons (🥇🥈🥉)
- Score values
- Winner highlighted (gold background)

**Props:**
- `consensusDecision`: Consensus object with all scores

**Interaction:**
- Click on bar to expand ScoreBreakdown
- Tooltip on hover with details

---

### ScoreBreakdown.tsx

**Purpose:**  
Detailed breakdown of scoring for one proposal.

**Layout:**
```
┌─────────────────────────────────────┐
│ gemini-2.5-pro Proposal - Score Breakdown │
├─────────────────────────────────────┤
│ Performance:     95.0 / 100  (50%)  │
│ ██████████████████████████████████  │
│                                     │
│ Storage:         95.0 / 100  (20%)  │
│ ██████████████████████████████████  │
│                                     │
│ Complexity:      85.0 / 100  (20%)  │
│ ████████████████████████████░░░░░░  │
│                                     │
│ Risk:            95.0 / 100  (10%)  │
│ ██████████████████████████████████  │
├─────────────────────────────────────┤
│ Weighted Total:  93.0 / 100         │
└─────────────────────────────────────┘
```

**Features:**
- Individual criterion scores
- Weight percentages shown
- Visual bars for each criterion
- Weighted total prominently displayed

**Props:**
- `agentType`: Which agent's scores
- `scores`: Score object from consensus

**Color Coding:**
- 90-100: Green (excellent)
- 70-89: Blue (good)
- 50-69: Yellow (fair)
- <50: Red (poor)

---

### DecisionRationale.tsx

**Purpose:**  
Display human-readable explanation of consensus decision.

**Content:**
```
┌─────────────────────────────────────┐
│ Why gemini-2.5-pro Won │
├─────────────────────────────────────┤
│ gemini-2.5-pro partial index proposal │
│ selected as optimal solution.       │
│                                     │
│ Performance: 82.6% improvement      │
│ (2.30s → 0.40s)                     │
│                                     │
│ Key Strengths:                      │
│ • Minimal storage overhead (12MB)   │
│ • Low operational complexity        │
│ • Easy rollback path                │
│                                     │
│ Runner-up Analysis:                 │
│ gemini-2.5-flash materialized view achieved │
│ higher performance (93.5%) but...   │
└─────────────────────────────────────┘
```

**Formatting:**
- Markdown rendering (bold, lists, etc.)
- Syntax highlighting for metrics
- Collapsible sections for long text
- Clean typography

**Props:**
- `rationale`: String (markdown formatted)

**Rendering:**
- Use markdown parser (react-markdown)
- Custom components for emphasis
- Line breaks preserved

---

### WinnerAnnouncement.tsx

**Purpose:**  
Celebratory display of winning proposal.

**Layout:**
```
┌─────────────────────────────────────┐
│           🏆 Winner 🏆             │
│                                     │
│      gemini-2.5-pro Partial Index │
│                                     │
│      82.6% Improvement              │
│      Score: 93.0 / 100              │
│                                     │
│ ✓ Optimization Applied Successfully │
└─────────────────────────────────────┘
```

**Animation:**
- Fade in on render
- Trophy icon pulse
- Confetti effect (optional, subtle)

**Props:**
- `proposal`: Winning proposal
- `applied`: Boolean (if already applied)

**States:**
- Winner selected but not applied
- Winner applied successfully
- Winner failed to apply (show error)

---

## 🧱 Common Components

### Button.tsx

**Purpose:**  
Reusable button with consistent styling.

**Variants:**
- Primary: Blue background, white text
- Secondary: White background, blue text, border
- Danger: Red background, white text
- Ghost: Transparent, text only

**Sizes:**
- Small: Compact padding
- Medium: Default
- Large: Prominent

**States:**
- Default
- Hover: Slight color change
- Active: Pressed state
- Disabled: Grayed out, no interaction
- Loading: Spinner icon, disabled

**Props:**
- `variant`: ButtonVariant enum
- `size`: ButtonSize enum
- `disabled`: Boolean
- `loading`: Boolean
- `onClick`: Click handler
- `children`: Button text/content

**Accessibility:**
- Keyboard navigation (Tab, Enter, Space)
- ARIA labels
- Focus indicators

---

### Card.tsx

**Purpose:**  
Container component with consistent styling.

**Features:**
- White background (or dark mode variant)
- Rounded corners
- Shadow (elevation)
- Padding

**Variants:**
- Default: Standard card
- Elevated: Larger shadow
- Outlined: Border instead of shadow
- Interactive: Hover effects (for clickable cards)

**Props:**
- `variant`: CardVariant
- `padding`: Spacing size
- `onClick`: Optional (makes card clickable)
- `children`: Card content

---

### Badge.tsx

**Purpose:**  
Small label for status, tags, or metadata.

**Use Cases:**
- Status indicators (Pending, Completed)
- Tags (SQL, Index, etc.)
- Counts (3 proposals)
- Metrics (82% improvement)

**Colors:**
- Gray: Neutral
- Blue: Info
- Green: Success
- Yellow: Warning
- Red: Error

**Sizes:**
- Small: Compact
- Medium: Default
- Large: Prominent

**Props:**
- `color`: BadgeColor
- `size`: BadgeSize
- `children`: Badge text

---

### Spinner.tsx

**Purpose:**  
Loading indicator.

**Variants:**
- Circular: Rotating circle
- Dots: Three pulsing dots
- Bars: Loading bars

**Sizes:**
- Small: 16px
- Medium: 24px
- Large: 48px

**Props:**
- `size`: SpinnerSize
- `variant`: SpinnerVariant

**Usage:**
- Page loading
- Button loading state
- Data fetching indicator
- Inline loading (small variant)

---

### Modal.tsx

**Purpose:**  
Overlay dialog for focused interactions.

**Features:**
- Backdrop (darkened background)
- Center-aligned content
- Close button (X icon)
- Escape key to close
- Click outside to close (optional)

**Props:**
- `isOpen`: Boolean
- `onClose`: Close callback
- `title`: Modal title
- `children`: Modal content
- `size`: "small" | "medium" | "large"

**Accessibility:**
- Focus trap (Tab stays within modal)
- ARIA role="dialog"
- Focus on first interactive element
- Return focus on close

---

### EmptyState.tsx

**Purpose:**  
Friendly message when no data to display.

**Layout:**
```
┌─────────────────────────────────────┐
│                                     │
│         [Illustration]              │
│                                     │
│     No tasks yet                    │
│     Create your first optimization! │
│                                     │
│     [Create Task] button            │
│                                     │
└─────────────────────────────────────┘
```

**Props:**
- `icon`: Illustration or icon
- `title`: Main message
- `description`: Supporting text
- `action`: Optional button/link

**Use Cases:**
- Empty task list
- No search results
- No proposals yet
- Failed to load (with retry action)

---

## 🎣 Custom Hooks

### useTasks.ts

**Purpose:**  
Manage task data fetching and mutations.

**Returns:**
```
{
  tasks: Task[]
  isLoading: boolean
  error: Error | null
  createTask: (data) => Promise<Task>
  refetch: () => void
}
```

**Implementation:**
- Uses React Query `useQuery` for fetching
- Uses `useMutation` for creating
- Cache key: ['tasks', filters]
- Auto-refetch every 30 seconds
- Optimistic updates on create

**Usage:**
```
In TaskList component:
  const { tasks, isLoading } = useTasks({ status: 'completed' })
  
In TaskSubmission component:
  const { createTask } = useTasks()
  await createTask(formData)
```

---

### useTask.ts

**Purpose:**  
Fetch single task by ID with real-time updates.

**Returns:**
```
{
  task: Task | null
  isLoading: boolean
  error: Error | null
  refetch: () => void
}
```

**Implementation:**
- React Query `useQuery`
- Cache key: ['task', taskId]
- WebSocket integration for real-time updates
- Refetch on WebSocket event

**Usage:**
```
In TaskDetail component:
  const { task, isLoading } = useTask(taskId)
```

---

### useAgents.ts

**Purpose:**  
Fetch all agents with status.

**Returns:**
```
{
  agents: Agent[]
  isLoading: boolean
  error: Error | null
}
```

**Implementation:**
- React Query `useQuery`
- Cache key: ['agents']
- Poll every 10 seconds (status changes frequently)
- Stale time: 5 seconds

---

### useOptimizations.ts

**Purpose:**  
Fetch proposals and benchmarks for a task.

**Returns:**
```
{
  proposals: Proposal[]
  benchmarks: Record<proposalId, BenchmarkResult[]>
  isLoading: boolean
}
```

**Implementation:**
- Parallel queries for proposals and benchmarks
- React Query for both
- Cache keys: ['proposals', taskId], ['benchmarks', proposalId]

---

### useConsensus.ts

**Purpose:**  
Fetch consensus decision for a task.

**Returns:**
```
{
  consensus: ConsensusDecision | null
  isLoading: boolean
  error: Error | null
}
```

**Implementation:**
- React Query `useQuery`
- Cache key: ['consensus', taskId]
- Enabled only if task status is completed
- WebSocket updates when consensus reached

---

### useWebSocket.ts

**Purpose:**  
Manage WebSocket connection and event handling.

**Returns:**
```
{
  isConnected: boolean
  subscribe: (taskId, callback) => void
  unsubscribe: (taskId) => void
}
```

**Implementation:**
- Single WebSocket connection per app instance
- Event routing to subscribed callbacks
- Automatic reconnection on disconnect
- Cleanup on unmount

**Event Handling:**
```
Event arrives:
  1. Parse JSON
  2. Extract taskId
  3. Find all subscribed callbacks for taskId
  4. Call each callback with event data
  5. Trigger React Query refetch for affected data
```

**Usage:**
```
In TaskDetail:
  const { subscribe, unsubscribe } = useWebSocket()
  
  useEffect(() => {
    subscribe(taskId, handleEvent)
    return () => unsubscribe(taskId)
  }, [taskId])
  
  function handleEvent(event) {
    if (event.type === 'proposal_submitted') {
      // Trigger refetch or update state
    }
  }
```

---

## 📦 State Management

### Server State (React Query)

**Purpose:**  
Manage data from API (tasks, agents, proposals, etc.)

**Benefits:**
- Automatic caching
- Background refetching
- Optimistic updates
- Request deduplication
- Pagination support

**Query Keys:**
```
['tasks'] - All tasks
['tasks', { status: 'completed' }] - Filtered tasks
['task', taskId] - Single task
['agents'] - All agents
['proposals', taskId] - Proposals for task
['consensus', taskId] - Consensus for task
```

**Configuration:**
```
QueryClient settings:
  defaultOptions: {
    queries: {
      staleTime: 30000 (30 seconds)
      cacheTime: 300000 (5 minutes)
      refetchOnWindowFocus: true
      retry: 2
    }
  }
```

---

### UI State (Context API)

**ThemeContext:**
- Light/Dark mode preference
- Persisted in localStorage
- Applies Tailwind dark classes

**NotificationContext:**
- Toast notifications
- Success/Error/Info messages
- Auto-dismiss after 5 seconds
- Stack multiple notifications

**Usage:**
```
In App.tsx:
  <ThemeProvider>
    <NotificationProvider>
      <Router />
    </NotificationProvider>
  </ThemeProvider>

In component:
  const { theme, toggleTheme } = useTheme()
  const { showNotification } = useNotification()
  
  showNotification({
    type: 'success',
    message: 'Task created successfully'
  })
```

---

### Component State (useState)

**Purpose:**  
Local component state that doesn't need sharing.

**Use Cases:**
- Form input values
- UI toggles (expanded/collapsed)
- Modal open/closed
- Temporary data

**Guidelines:**
- Keep state as local as possible
- Lift up only when multiple components need it
- Derive state from props when possible
- Avoid duplicating server state

---

## 🛣️ Routing

### React Router Setup

**Routes:**
```
/ - HomePage (dashboard)
/tasks - TaskListPage
/tasks/new - TaskSubmissionPage
/tasks/:id - TaskDetailPage
/agents - AgentsPage
/stats - StatsPage (optional)
```

**Route Configuration:**
```
Using React Router v6:
  <Routes>
    <Route path="/" element={<HomePage />} />
    <Route path="/tasks" element={<TaskListPage />} />
    <Route path="/tasks/new" element={<TaskSubmissionPage />} />
    <Route path="/tasks/:id" element={<TaskDetailPage />} />
    <Route path="/agents" element={<AgentsPage />} />
    <Route path="*" element={<NotFoundPage />} />
  </Routes>
```

**Navigation:**
- Programmatic: `useNavigate()` hook
- Declarative: `<Link to="/tasks">` component
- URL params: `useParams()` hook

---

### Protected Routes (Future)

**Authentication:**
```
If implementing auth:
  <Route 
    path="/tasks/new" 
    element={
      <RequireAuth>
        <TaskSubmissionPage />
      </RequireAuth>
    } 
  />

RequireAuth component:
  - Check if user authenticated
  - If not, redirect to login
  - If yes, render children
```

---

## ⚡ Real-Time Updates

### WebSocket Integration

**Connection Management:**

**Establish Connection:**
```
On app mount:
  1. Create WebSocket connection
  2. Handle 'connection_established' event
  3. Set isConnected state
  4. Attach event listeners
```

**Event Routing:**
```
When event arrives:
  1. Parse JSON
  2. Determine event type
  3. Route to appropriate handler
  4. Update React Query cache
  5. Trigger component re-renders
```

---

### Event Handlers

**task_created:**
```
Action: Add to tasks cache (prepend to list)
UI: Show toast notification
```

**agents_assigned:**
```
Action: Update task in cache
UI: Update progress timeline in TaskDetail
```

**proposal_submitted:**
```
Action: Add to proposals cache
UI: Show new proposal in ProposalList
```

**consensus_reached:**
```
Action: Update consensus cache
UI: Show winner announcement, confetti
```

**task_completed:**
```
Action: Update task status in cache
UI: Show completion notification, update badge
```

---

### Optimistic Updates

**Example: Create Task**

```
Flow:
  1. User submits form
  2. Immediately add task to UI (optimistic)
     - Status: "pending"
     - ID: temporary (temp-123)
  3. Send POST request
  4. On success:
     - Replace temp task with real task
     - Update with real ID
  5. On error:
     - Remove optimistic task
     - Show error message
     - Revert UI state
```

**Benefits:**
- Instant feedback
- Perceived performance
- Better UX

**Risks:**
- Must handle rollback on error
- Consistent state management required

---

## 🎯 Summary

This frontend architecture provides:

**Component Organization:**
- Feature-based structure (task, agent, optimization, consensus)
- 20+ components across 5 feature domains
- Atomic design principles
- Reusable common components

**Custom Hooks:**
- 6 data hooks (useTasks, useTask, useAgents, etc.)
- 1 WebSocket hook
- React Query integration
- Real-time data synchronization

**State Management:**
- Server state via React Query (caching, refetching)
- UI state via Context API (theme, notifications)
- Component state via useState (local)
- Clear separation of concerns

**Real-Time Updates:**
- WebSocket connection management
- Event routing and handling
- Optimistic updates
- Cache synchronization

**Key Technologies:**
- React 18 + TypeScript
- Vite (build tool)
- Tailwind CSS (styling)
- React Query (server state)
- React Router (routing)
- Native WebSocket (real-time)

**UI/UX Features:**
- Responsive design (mobile-first)
- Dark mode support
- Loading states
- Error handling
- Empty states
- Accessibility (ARIA, keyboard nav)

---

**Related Documentation:**
- Previous: [08-API-SPECIFICATION.md](08-API-SPECIFICATION.md) 
  - API endpoints that components consume
- See also: [01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md) 
  - User flows that UI implements
- See also: [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md) 
  - Overall system context

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```
