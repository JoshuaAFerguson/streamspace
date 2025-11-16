# StreamSpace UI Tests

This directory contains test utilities and setup for the StreamSpace UI.

## Test Infrastructure

- **Framework**: Vitest (fast, modern test runner for Vite projects)
- **React Testing**: `@testing-library/react` for component testing
- **DOM Matchers**: `@testing-library/jest-dom` for enhanced assertions
- **User Interactions**: `@testing-library/user-event` for simulating user actions

## Running Tests

```bash
# Run tests in watch mode (development)
npm test

# Run tests once (CI)
npm run test:run

# Run tests with UI dashboard
npm run test:ui

# Generate coverage report
npm run test:coverage
```

## Writing Tests

### Component Test Example

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import MyComponent from './MyComponent';

describe('MyComponent', () => {
  it('renders correctly', () => {
    render(<MyComponent title="Test" />);
    expect(screen.getByText('Test')).toBeInTheDocument();
  });

  it('handles click events', () => {
    const onClick = vi.fn();
    render(<MyComponent onClick={onClick} />);

    fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
```

### Hook Test Example

```typescript
import { renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import useApi from './useApi';

describe('useApi', () => {
  it('fetches data successfully', async () => {
    const { result } = renderHook(() => useApi('/api/sessions'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBeDefined();
  });
});
```

## Test Organization

Tests should be colocated with their source files:

```
src/
├── components/
│   ├── SessionCard.tsx
│   └── SessionCard.test.tsx
├── hooks/
│   ├── useApi.ts
│   └── useApi.test.ts
└── pages/
    ├── Dashboard.tsx
    └── Dashboard.test.tsx
```

## Coverage Goals

- **Overall**: 80%+
- **Critical paths** (auth, session management): 90%+
- **UI components**: 80%+
- **Hooks and utilities**: 85%+

## Best Practices

1. **Test behavior, not implementation** - Focus on user-facing behavior
2. **Use semantic queries** - Prefer `getByRole`, `getByLabelText` over `getByTestId`
3. **Keep tests focused** - One test should verify one behavior
4. **Mock external dependencies** - API calls, browser APIs, etc.
5. **Use realistic data** - Test with data similar to production
6. **Test accessibility** - Verify keyboard navigation, ARIA attributes
7. **Avoid implementation details** - Don't test internal state or private methods

## Troubleshooting

### Tests fail to find elements

- Use `screen.debug()` to see the rendered output
- Check that components are properly imported and rendered
- Verify queries are correct (`getBy*`, `findBy*`, `queryBy*`)

### Async operations timeout

- Use `waitFor` or `findBy*` queries for async operations
- Increase timeout if needed: `waitFor(() => {...}, { timeout: 5000 })`

### Mock not working

- Ensure mocks are defined before imports: `vi.mock('./module')`
- Clear mocks between tests: `afterEach(() => { vi.clearAllMocks() })`

## Resources

- [Vitest Documentation](https://vitest.dev/)
- [Testing Library Docs](https://testing-library.com/docs/react-testing-library/intro/)
- [Common Testing Mistakes](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)
