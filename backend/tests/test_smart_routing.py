import pytest
from local_intelligence import get_local_intelligence

@pytest.fixture(scope="module")
def li():
    # Load the local intelligence instance once for all tests
    return get_local_intelligence()

def test_smart_routing_categories(li):
    """
    E2E Automation test for the Smart Context Pipeline (P36-T2).
    This tests the SmolLM2 zero-shot classification against all expected categories.
    """
    
    test_cases = [
        # 1. Conceptual
        ("What is the Virtual DOM in React, and how does it improve performance?", "conceptual"),
        ("Explain the difference between useMemo and useCallback.", "conceptual"),
        
        # 2. Coding
        ("Write a custom React hook called useDebounce that takes a value and a delay.", "coding"),
        ("Implement a simple React component that fetches and displays a list of users from an API.", "coding"),
        
        # 3. System Design
        ("How would you design the frontend architecture for a real-time collaborative document editor like Google Docs using React?", "system_design"),
        ("Design a highly scalable e-commerce product listing page with infinite scrolling.", "system_design"),
        
        # 4. Opinion
        ("Do you prefer using Redux Toolkit, Zustand, or React Context for global state management, and why?", "opinion"),
        ("What is your opinion on using Tailwind CSS versus CSS-in-JS libraries?", "opinion"),
        
        # 5. Behavioral
        ("Tell me about a time you had to optimize a severely underperforming React application. What did you do?", "behavioral"),
        ("Describe a situation where you disagreed with a senior engineer about the architecture of a React component.", "behavioral"),
        
        # 6. Noise (Control Test)
        ("Hmm, okay.", "noise"),
        ("That makes sense, thank you.", "noise")
    ]
    
    failed_cases = []
    
    for question, expected_category in test_cases:
        actual_category = li.classify_question(question)
        if actual_category != expected_category:
            failed_cases.append(f"Failed '{question[:30]}...': Expected '{expected_category}', got '{actual_category}'")
            
    # Assert all passed
    assert not failed_cases, "\n".join(failed_cases)
