import google.generativeai as genai
import os
from dotenv import load_dotenv
import logging
from typing import List, Dict, Optional, Tuple

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)
MODEL_NAME = "gemini-2.0-flash"
gemini_model = None
try:
    load_dotenv()
    api_key = os.getenv("GEMINI_API_KEY")
    if not api_key:
        raise ValueError("GEMINI_API_KEY not found.")
    genai.configure(api_key=api_key)
    gemini_model = genai.GenerativeModel(MODEL_NAME)
    logger.info(f"Gemini model '{MODEL_NAME}' initialized successfully.")
except Exception as e:
    logger.error(f"Error during initialization: {e}", exc_info=True)

def generate_chat_response(
    prompt: str,
    history: Optional[List[Dict]] = None
) -> Tuple[Optional[str], List[Dict]]:
    """
    Generates a chat response using the configured Gemini model and
    returns the response along with the *updated* history list.

    Args:
        prompt (str): The user's latest message.
        history (Optional[List[Dict]], optional): The conversation history *before*
                                                  this turn. Defaults to None (starts fresh).
                                                  Format: [{'role': 'user'|'model', 'parts': [str]}]
                                                  This list is *not* modified.

    Returns:
        Tuple[Optional[str], List[Dict]]: A tuple containing:
            - The generated text response (str) or None if an error occurs.
            - The new, updated history list (List[Dict]), including the
              latest user prompt and model response. If an error occurred,
              the returned list is the same as the input history.
    """
    if not gemini_model:
        logger.error("Gemini model is not initialized.")
        return None, history or []
    current_history_dicts = [item.model_dump() if hasattr(item, 'model_dump') else item for item in history] if history else []
    updated_history = list(current_history_dicts)

    try:
        logger.info(f"Starting chat session. History length: {len(current_history_dicts)}. User prompt: '{prompt[:100]}...'")
        chat_session = gemini_model.start_chat(history=current_history_dicts)
        response = chat_session.send_message(prompt)
        response_text = response.text

        updated_history.append({'role': 'user', 'parts': [prompt]})
        updated_history.append({'role': 'model', 'parts': [response_text]})

        logger.info(f"Received response. Returning updated history (list of dicts). New length: {len(updated_history)}")
        return response_text, updated_history

    except genai.types.StopCandidateException as sce:
        logger.warning(f"Generation stopped: {sce}")
        partial_text = "Generation stopped due to safety or other reasons."
        try:
            if response and hasattr(response, 'text'):
                 partial_text = response.text
        except NameError:
            pass
        return partial_text, current_history_dicts

    except Exception as e:
        logger.error(f"Error during Gemini API call: {e}", exc_info=True)
        return None, current_history_dicts

async def generate_trade_summary(trade_data: dict, strategy_profile: dict, chat_context: str = ""):
    """
    Generates a trade summary using the Gemini model.
    (Placeholder - needs prompt construction)
    """
    if not gemini_model:
        logger.error("Gemini model not initialized.")
        return "Error: AI model not available."
    if not api_key:
         logger.error("Gemini API key not configured.")
         return "Error: AI Service not configured."

    prompt = f"""
    Analyze the following trade based on the provided strategy profile and conversation context.

    **Trade Data:**
    {trade_data}

    **Strategy Profile:**
    {strategy_profile}

    **Conversation Context:**
    {chat_context}

    **Instructions:**
    Create a summary with:
    - Setup quality rating (A/B/C)
    - Emotional drivers (e.g., FOMO, Patience, Discipline)
    - Mistakes made (if any)
    - Key lessons learned
    - Specific suggestions for improvement
    """
    logger.info("Generating trade summary...")
    try:
        response = await gemini_model.generate_content_async(prompt)
        logger.info("Trade summary generated.")
        return response.text
    except Exception as e:
        logger.error(f"Error generating trade summary: {e}")
        return f"Error communicating with AI: {e}"


async def suggest_trade_tags(trade_data: dict, notes: str):
    """
    Suggests relevant tags for a trade using the Gemini model.
    (Placeholder - needs prompt construction and tag list)
    """
    if not gemini_model:
        logger.error("Gemini model not initialized.")
        return ["Error: AI model not available."]
    if not api_key:
         logger.error("Gemini API key not configured.")
         return ["Error: AI Service not configured."]

    allowed_tags = ["FOMO", "Revenge Trading", "Impulsive Entry", "Good Patience", "Discipline", "Hesitation", "Over-Leveraged", "Cut Winner Short", "Let Loser Run"] # Example list
    prompt = f"""
    Based on the following trade data and notes, suggest 1-3 concise tags summarizing the execution, behavior, or psychology.

    **Trade Data:**
    {trade_data}

    **User Notes:**
    {notes}

    **Instructions:**
    - Suggest only tags from this allowed list: {', '.join(allowed_tags)}
    - If no specific behavior/psychology is evident, suggest 'Standard Trade'.
    - Output the suggested tags as a comma-separated list.
    """
    logger.info("Suggesting trade tags...")
    try:
        response = await gemini_model.generate_content_async(prompt)
        logger.info("Trade tags suggested.")
        suggested_tags = [tag.strip() for tag in response.text.split(',') if tag.strip()]
        valid_suggestions = [tag for tag in suggested_tags if tag in allowed_tags or tag == 'Standard Trade']
        return valid_suggestions if valid_suggestions else ["Standard Trade"] # Default if parsing fails or no valid tags found
    except Exception as e:
        logger.error(f"Error suggesting trade tags: {e}")
        return [f"Error communicating with AI: {e}"]
